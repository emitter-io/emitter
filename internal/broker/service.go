/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/

package broker

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/emitter-io/address"
	"github.com/emitter-io/emitter/internal/broker/cluster"
	"github.com/emitter-io/emitter/internal/broker/keygen"
	"github.com/emitter-io/emitter/internal/broker/survey"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/network/listener"
	"github.com/emitter-io/emitter/internal/network/websocket"
	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/provider/monitor"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/security/license"
	"github.com/emitter-io/stats"
	"github.com/kelindar/tcp"
)

// Service represents the main structure.
type Service struct {
	connections   int64                // The number of currently open connections.
	context       context.Context      // The context for the service.
	cancel        context.CancelFunc   // The cancellation function.
	License       license.License      // The licence for this emitter server.
	Keygen        *keygen.Provider     // The key generation provider.
	Config        *config.Config       // The configuration for the service.
	subscriptions *message.Trie        // The subscription matching trie.
	http          *http.Server         // The underlying HTTP server.
	tcp           *tcp.Server          // The underlying TCP server.
	cluster       *cluster.Swarm       // The gossip-based cluster mechanism.
	presence      chan *presenceNotify // The channel for presence notifications.
	surveyor      *survey.Surveyor     // The generic query manager.
	contracts     contract.Provider    // The contract provider for the service.
	storage       storage.Storage      // The storage provider for the service.
	monitor       monitor.Storage      // The storage provider for stats.
	measurer      stats.Measurer       // The monitoring registry for the service.
	metering      usage.Metering       // The usage storage for metering contracts.
}

// NewService creates a new service.
func NewService(ctx context.Context, cfg *config.Config) (s *Service, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	s = &Service{
		context:       ctx,
		cancel:        cancel,
		Config:        cfg,
		subscriptions: message.NewTrie(),
		http:          new(http.Server),
		tcp:           new(tcp.Server),
		presence:      make(chan *presenceNotify, 100),
		storage:       new(storage.Noop),
		measurer:      stats.New(),
	}

	// Create a new HTTP request multiplexer
	mux := http.NewServeMux()

	// Attach handlers
	s.http.Handler = mux
	s.tcp.OnAccept = s.onAcceptConn

	// Parse the license
	if s.License, err = license.Parse(cfg.License); err != nil {
		return nil, err
	}

	// Create a new cluster if we have this configured
	if cfg.Cluster != nil {
		s.cluster = cluster.NewSwarm(cfg.Cluster)
		s.cluster.OnMessage = s.onPeerMessage
		s.cluster.OnSubscribe = s.Subscribe
		s.cluster.OnUnsubscribe = s.Unsubscribe
	}

	// Load the logging provider
	logging.Logger = config.LoadProvider(cfg.Logging, logging.NewStdErr()).(logging.Logging)
	logging.LogTarget("service", "configured logging provider", logging.Logger.Name())

	// Load the storage provider
	ssdstore := storage.NewSSD(s)
	memstore := storage.NewInMemory(s)
	s.storage = config.LoadProvider(cfg.Storage, storage.NewNoop(), memstore, ssdstore).(storage.Storage)
	logging.LogTarget("service", "configured message storage", s.storage.Name())

	// Load the metering provider
	s.metering = config.LoadProvider(cfg.Metering, usage.NewNoop(), usage.NewHTTP()).(usage.Metering)
	logging.LogTarget("service", "configured usage metering", s.metering.Name())

	// Load the contract provider
	s.contracts = config.LoadProvider(cfg.Contract,
		contract.NewSingleContractProvider(s.License, s.metering),
		contract.NewHTTPContractProvider(s.License, s.metering)).(contract.Provider)
	logging.LogTarget("service", "configured contracts provider", s.contracts.Name())

	// Load the monitor storage provider
	nodeName := address.Fingerprint(s.ID()).String()
	sampler := newSampler(s, s.measurer)
	s.monitor = config.LoadProvider(cfg.Monitor,
		monitor.NewSelf(sampler, s.selfPublish),
		monitor.NewNoop(),
		monitor.NewHTTP(sampler),
		monitor.NewStatsd(sampler, nodeName),
		monitor.NewPrometheus(sampler, mux),
	).(monitor.Storage)
	logging.LogTarget("service", "configured monitoring sink", s.monitor.Name())

	// Attach survey handlers
	s.surveyor = survey.New(s, s.cluster)
	if s.cluster != nil {
		s.surveyor.HandleFunc(s, ssdstore, memstore)
	}

	// Create a new cipher from the licence provided
	cipher, err := s.License.Cipher()
	if err != nil {
		return nil, err
	}

	// Attach handlers
	s.Keygen = keygen.NewProvider(cipher, s.contracts)
	if cfg.Debug {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	mux.HandleFunc("/health", s.onHealth)
	mux.HandleFunc("/keygen", s.Keygen.HTTP())
	mux.HandleFunc("/presence", s.onHTTPPresence)
	mux.HandleFunc("/", s.onRequest)

	// Addresses and things
	logging.LogTarget("service", "configured node name", nodeName)
	return s, nil
}

// ID returns the local node ID.
func (s *Service) ID() uint64 {
	if s.cluster != nil {
		return s.cluster.ID()
	}

	return uint64(address.GetHardware())
}

// NumPeers returns the number of peers of this service.
func (s *Service) NumPeers() int {
	if s.cluster != nil {
		return s.cluster.NumPeers()
	}

	return 0
}

// Listen starts the service.
func (s *Service) Listen() (err error) {
	defer s.Close()
	s.hookSignals()
	s.pollPresenceChange()

	// Create the cluster if required
	if s.cluster != nil {
		if s.cluster.Listen(s.context); err != nil {
			panic(err)
		}

		// Join our seed
		s.Join(s.Config.Cluster.Seed)

		// Subscribe to the query channel
		s.surveyor.Start()
	}

	// Setup the listeners on both default and a secure addresses
	s.listen(s.Config.Addr(), nil)
	if tls, tlsValidator, ok := s.Config.Certificate(); ok {

		// If we need to validate certificate, spin up a listener on port 80
		// More info: https://community.letsencrypt.org/t/2018-01-11-update-regarding-acme-tls-sni-and-shared-hosting-infrastructure/50188
		if tlsValidator != nil {
			logging.LogAction("service", "exposing autocert TLS validation on :80")
			go http.ListenAndServe(":80", tlsValidator)
		}

		if tlsAddr, err := address.Parse(s.Config.TLS.ListenAddr, 443); err == nil {
			s.listen(tlsAddr, tls)
		}
	}

	// Block
	logging.LogAction("service", "service started")
	select {}
}

// listen configures an main listener on a specified address.
func (s *Service) listen(addr *net.TCPAddr, conf *tls.Config) {

	// Create new listener
	logging.LogTarget("service", "starting the listener", addr)
	l, err := listener.New(addr.String(), listener.Config{
		FlushRate: s.Config.Limit.FlushRate,
		TLS:       conf,
	})
	if err != nil {
		panic(err)
	}

	// Set the read timeout on our mux listener
	l.SetReadTimeout(120 * time.Second)

	// Configure the matchers
	l.ServeAsync(listener.MatchHTTP(), s.http.Serve)
	l.ServeAsync(listener.MatchAny(), s.tcp.Serve)
	go l.Serve()
}

// Join attempts to join a set of existing peers.
func (s *Service) Join(peers ...string) []error {
	return s.cluster.Join(peers...)
}

// notifyPresenceChange sends out an event to notify when a client is subscribed/unsubscribed.
func (s *Service) pollPresenceChange() {
	go func() {
		for {
			select {
			case <-s.context.Done():
				return
			case notif := <-s.presence:
				s.notifyPresenceEvent(notif, nil)
			}
		}
	}()
}

// notifyPresenceChange sends out an event to notify when a client is subscribed/unsubscribed.
func (s *Service) notifyPresenceEvent(ev *presenceNotify, filter func(message.Subscriber) bool) {
	channel := []byte("emitter/presence/") // TODO: avoid allocation
	if encoded, ok := ev.Encode(); ok {
		s.Publish(message.New(ev.Ssid, channel, encoded), filter)
	}
}

// NotifySubscribe notifies the swarm when a subscription occurs.
func (s *Service) notifySubscribe(ev *event.Subscription) {

	// If we have a new direct subscriber, issue presence message and publish it
	if ev.Channel != nil {
		s.presence <- newPresenceNotify(presenceSubscribeEvent, ev)
	}

	// Notify our cluster that the client just subscribed.
	if s.cluster != nil {
		s.cluster.NotifySubscribe(ev)
	}
}

// NotifyUnsubscribe notifies the swarm when an unsubscription occurs.
func (s *Service) notifyUnsubscribe(ev *event.Subscription) {

	// If we have a new direct subscriber, issue presence message and publish it
	if ev.Channel != nil {
		s.presence <- newPresenceNotify(presenceUnsubscribeEvent, ev)
	}

	// Notify our cluster that the client just unsubscribed.
	if s.cluster != nil {
		s.cluster.NotifyUnsubscribe(ev)
	}
}

// Occurs when a new client connection is accepted.
func (s *Service) onAcceptConn(t net.Conn) {
	conn := s.newConn(t, s.Config.Limit.ReadRate)
	go conn.Process()
}

// Occurs when a new HTTP request is received.
func (s *Service) onRequest(w http.ResponseWriter, r *http.Request) {
	if ws, ok := websocket.TryUpgrade(w, r); ok {
		s.onAcceptConn(ws)
		return
	}
}

// Occurs when a new HTTP health check is received.
func (s *Service) onHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

// Occurs when a new HTTP presence request is received.
func (s *Service) onHTTPPresence(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Deserialize the body.
	msg := presenceRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Attempt to parse the key, this should be a master key
	key, err := s.Keygen.DecryptKey(msg.Key)
	if err != nil || !key.HasPermission(security.AllowPresence) || key.IsExpired() {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := s.contracts.Get(key.Contract())
	if !contractFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Validate the contract
	if !contract.Validate(key) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Ensure we have trailing slash
	if !strings.HasSuffix(msg.Channel, "/") {
		msg.Channel = msg.Channel + "/"
	}

	// Parse the channel
	channel := security.ParseChannel([]byte("emitter/" + msg.Channel))
	if channel.ChannelType == security.ChannelInvalid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create the ssid for the presence
	ssid := message.NewSsid(key.Contract(), channel.Query)
	now := time.Now().UTC().Unix()
	who := getAllPresence(s, ssid)
	resp, err := json.Marshal(&presenceResponse{
		Time:    now,
		Event:   presenceStatusEvent,
		Channel: msg.Channel,
		Who:     who,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(resp)
	return
}

// Subscribe subscribes to a channel.
func (s *Service) Subscribe(sub message.Subscriber, ev *event.Subscription) bool {
	if _, err := s.subscriptions.Subscribe(ev.Ssid, sub); err != nil {
		return false // Unable to subscribe
	}

	// Broadcast direct subscriptions
	if sub.Type() == message.SubscriberDirect {
		s.notifySubscribe(ev)
	}
	return true
}

// Unsubscribe unsubscribes from a channel
func (s *Service) Unsubscribe(sub message.Subscriber, ev *event.Subscription) (ok bool) {
	subscribers := s.subscriptions.Lookup(ev.Ssid, nil)
	if ok = subscribers.Contains(sub); ok {
		s.subscriptions.Unsubscribe(ev.Ssid, sub)
	}

	// Broadcast direct unsubscriptions
	if sub.Type() == message.SubscriberDirect {
		s.notifyUnsubscribe(ev)
	}

	// If the peer is offline, notify the presence
	if sub.Type() == message.SubscriberOffline {
		s.notifyPresenceEvent(newPresenceNotify(presenceUnsubscribeEvent, ev), func(s message.Subscriber) bool {
			return s.Type() == message.SubscriberDirect
		})
	}

	return
}

// Occurs when a message is received from a peer.
func (s *Service) onPeerMessage(m *message.Message) {
	defer s.measurer.MeasureElapsed("peer.msg", time.Now())
	size, n := len(m.Payload), 0
	filter := func(s message.Subscriber) bool {
		return s.Type() == message.SubscriberDirect // only local subscribers
	}

	// Iterate through all subscribers and send them the message
	for _, subscriber := range s.subscriptions.Lookup(m.Ssid(), filter) {
		subscriber.Send(m)
		n += size
	}

	// Get the contract
	contract, contractFound := s.contracts.Get(m.Contract())
	if contractFound {
		contract.Stats().AddEgress(int64(n))
	}
}

// Survey is a mechanism where a message from one node is broadcasted to the
// entire cluster and each node in the group responds to the message.
func (s *Service) Survey(query string, payload []byte) (message.Awaiter, error) {
	if s.surveyor != nil {
		return s.surveyor.Query(query, payload)
	}

	return nil, errors.New("Query manager was not setup")
}

// Publish publishes a message to everyone and returns the number of outgoing bytes written.
func (s *Service) Publish(m *message.Message, filter func(message.Subscriber) bool) (n int64) {
	size := m.Size()

	for _, subscriber := range s.subscriptions.Lookup(m.Ssid(), filter) {
		subscriber.Send(m)
		if subscriber.Type() == message.SubscriberDirect {
			n += size
		}
	}
	return
}

// Authorize attempts to authorize a channel with its key
func (s *Service) authorize(channel *security.Channel, permission uint8) (contract.Contract, security.Key, bool) {

	// Attempt to parse the key
	key, err := s.Keygen.DecryptKey(string(channel.Key))
	if err != nil || key.IsExpired() {
		return nil, nil, false
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := s.contracts.Get(key.Contract())
	if !contractFound || !contract.Validate(key) || !key.HasPermission(permission) || !key.ValidateChannel(channel) {
		return nil, nil, false
	}

	// Return the contract and the key
	return contract, key, true
}

// SelfPublish publishes a message to itself.
func (s *Service) selfPublish(channelName string, payload []byte) {
	channel := security.ParseChannel([]byte("emitter/" + channelName))
	if channel.ChannelType == security.ChannelStatic {
		s.Publish(message.New(
			message.NewSsid(s.License.Contract(), channel.Query),
			channel.Channel,
			payload,
		), nil)
	}
}

// OnSignal will be called when a OS-level signal is received.
func (s *Service) onSignal(sig os.Signal) {
	switch sig {
	case syscall.SIGTERM:
		fallthrough
	case syscall.SIGINT:
		logging.LogAction("service", fmt.Sprintf("received signal %s, exiting...", sig.String()))
		s.Close()
		os.Exit(0)
	}
}

// OnSignal starts the signal processing and makes su
func (s *Service) hookSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range c {
			s.onSignal(sig)
		}
	}()
}

// Close closes gracefully the service.,
func (s *Service) Close() {
	if s.cancel != nil {
		s.cancel()
	}

	// Gracefully dispose all of our resources
	dispose(s.cluster)
	dispose(s.storage)

}

func dispose(resource io.Closer) {
	closable := !(resource == nil || (reflect.ValueOf(resource).Kind() == reflect.Ptr && reflect.ValueOf(resource).IsNil()))
	if closable {
		if err := resource.Close(); err != nil {
			logging.LogError("service", "error during close", err)
		}
	}
}
