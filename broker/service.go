/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
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
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emitter-io/emitter/broker/cluster"
	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/broker/storage"
	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/emitter/network/listener"
	"github.com/emitter-io/emitter/network/websocket"
	"github.com/emitter-io/emitter/security"
	"github.com/emitter-io/emitter/security/usage"
	"github.com/emitter-io/emitter/utils"
	"github.com/kelindar/tcp"
)

// Service represents the main structure.
type Service struct {
	Closing       chan bool                 // The channel for closing signal.
	Cipher        *security.Cipher          // The cipher to use for decoding and encoding keys.
	License       *security.License         // The licence for this emitter server.
	Config        *config.Config            // The configuration for the service.
	subscriptions *message.Trie             // The subscription matching trie.
	http          *http.Server              // The underlying HTTP server.
	tcp           *tcp.Server               // The underlying TCP server.
	cluster       *cluster.Swarm            // The gossip-based cluster mechanism.
	startTime     time.Time                 // The start time of the service.
	presence      chan *presenceNotify      // The channel for presence notifications.
	querier       *QueryManager             // The generic query manager.
	contracts     security.ContractProvider // The contract provider for the service.
	storage       storage.Storage           // The storage provider for the service.
	metering      usage.Metering            // The usage storage for metering contracts.
}

// NewService creates a new service.
func NewService(cfg *config.Config) (s *Service, err error) {
	s = &Service{
		Closing:       make(chan bool),
		Config:        cfg,
		subscriptions: message.NewTrie(),
		http:          new(http.Server),
		tcp:           new(tcp.Server),
		presence:      make(chan *presenceNotify, 100),
		storage:       new(storage.Noop),
	}

	// Create a new HTTP request multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.onHealth)
	mux.HandleFunc("/keygen", s.onHTTPKeyGen)
	mux.HandleFunc("/debug/pprof/", pprof.Index)          // TODO: use config flag to enable/disable this
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline) // TODO: use config flag to enable/disable this
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile) // TODO: use config flag to enable/disable this
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)   // TODO: use config flag to enable/disable this
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)     // TODO: use config flag to enable/disable this
	mux.HandleFunc("/", s.onRequest)

	// Attach handlers
	s.http.Handler = mux
	s.tcp.OnAccept = s.onAcceptConn
	s.querier = newQueryManager(s)

	// Parse the license
	if s.License, err = security.ParseLicense(cfg.License); err != nil {
		return nil, err
	}

	// Create a new cipher from the licence provided
	if s.Cipher, err = s.License.Cipher(); err != nil {
		return nil, err
	}

	// Create a new cluster if we have this configured
	if cfg.Cluster != nil {
		s.cluster = cluster.NewSwarm(cfg.Cluster, s.Closing)
		s.cluster.OnMessage = s.onPeerMessage
		s.cluster.OnSubscribe = s.onSubscribe
		s.cluster.OnUnsubscribe = s.onUnsubscribe

		// Attach query handlers
		s.querier.HandleFunc(s.onPresenceQuery)
	}

	// Load the logging provider
	logging.Logger = config.LoadProvider(cfg.Logging, logging.NewStdErr(), logging.NewLoggly()).(logging.Logging)
	logging.LogTarget("service", "configured logging provider", logging.Logger.Name())

	// Load the storage provider
	memstore := &storage.InMemory{Query: s.Query}
	s.querier.HandleFunc(memstore.OnRequest)
	s.storage = config.LoadProvider(cfg.Storage,
		new(storage.Noop),
		memstore).(storage.Storage)
	logging.LogTarget("service", "configured storage provider", s.storage.Name())

	// Load the metering provider
	s.metering = config.LoadProvider(cfg.Metering, usage.NewNoop(), usage.NewHTTP()).(usage.Metering)
	logging.LogTarget("service", "configured metering provider", s.metering.Name())

	// Load the contract provider
	s.contracts = config.LoadProvider(cfg.Contract,
		security.NewSingleContractProvider(s.License, s.metering),
		security.NewHTTPContractProvider(s.License, s.metering)).(security.ContractProvider)
	logging.LogTarget("service", "configured contracts provider", s.contracts.Name())

	// Addresses and things
	logging.LogTarget("service", "configured external address", address.External())
	logging.LogTarget("service", "configured node name", address.Fingerprint(s.LocalName()).String())
	return s, nil
}

// LocalName returns the local node name.
func (s *Service) LocalName() uint64 {
	if s.cluster != nil {
		return s.cluster.ID()
	}

	return uint64(address.Hardware())
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
	s.notifyPresenceChange()

	// Create the cluster if required
	if s.cluster != nil {
		if s.cluster.Listen(); err != nil {
			panic(err)
		}

		// Join our seed
		s.Join(s.Config.Cluster.Seed)

		// Subscribe to the query channel
		s.querier.Start()
	}

	// Setup the listeners on both default and a secure addresses
	s.listen(s.Config.ListenAddr)
	if cert, err := s.Config.TLS.Load(); err != nil {
		logging.LogError("service", "setup of TLS/SSL listener", err)
	} else {
		s.listen(s.Config.TLS.ListenAddr, cert)
	}

	// Set the start time and report status
	s.startTime = time.Now().UTC()
	utils.Repeat(s.reportStatus, 100*time.Millisecond, s.Closing)
	logging.LogAction("service", "service started")

	// Block
	select {}
}

// listen configures an main listener on a specified address.
func (s *Service) listen(address string, certs ...tls.Certificate) {
	logging.LogTarget("service", "starting the listener", address)
	l, err := listener.New(address, certs...)
	if err != nil {
		panic(err)
	}

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
func (s *Service) notifyPresenceChange() {
	go func() {
		channel := []byte("emitter/presence/")
		for {
			select {
			case <-s.Closing:
				return
			case notif := <-s.presence:
				if encoded, ok := notif.Encode(); ok {
					s.publish(notif.Ssid, channel, encoded)
				}
			}
		}
	}()
}

// NotifySubscribe notifies the swarm when a subscription occurs.
func (s *Service) notifySubscribe(conn *Conn, ssid message.Ssid, channel []byte) {

	// If we have a new direct subscriber, issue presence message and publish it
	if channel != nil {
		s.presence <- newPresenceNotify(ssid, presenceSubscribeEvent, string(channel), conn.ID(), conn.username)
	}

	// Notify our cluster that the client just subscribed.
	if s.cluster != nil {
		s.cluster.NotifySubscribe(conn.luid, ssid)
	}
}

// NotifyUnsubscribe notifies the swarm when an unsubscription occurs.
func (s *Service) notifyUnsubscribe(conn *Conn, ssid message.Ssid, channel []byte) {

	// If we have a new direct subscriber, issue presence message and publish it
	if channel != nil {
		s.presence <- newPresenceNotify(ssid, presenceUnsubscribeEvent, string(channel), conn.ID(), conn.username)
	}

	// Notify our cluster that the client just unsubscribed.
	if s.cluster != nil {
		s.cluster.NotifyUnsubscribe(conn.luid, ssid)
	}
}

// Occurs when a new client connection is accepted.
func (s *Service) onAcceptConn(t net.Conn) {
	conn := s.newConn(t)
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

// Occurs when a new HTTP request is received.
func (s *Service) onHTTPKeyGen(w http.ResponseWriter, r *http.Request) {
	if resp, err := http.Get("http://s3-eu-west-1.amazonaws.com/cdn.emitter.io/web/keygen.html"); err == nil {
		if content, err := ioutil.ReadAll(resp.Body); err == nil {
			w.Write(content)
			return
		}
	}
}

// Occurs when a peer has a new subscription.
func (s *Service) onSubscribe(ssid message.Ssid, sub message.Subscriber) bool {
	if _, err := s.subscriptions.Subscribe(ssid, sub); err != nil {
		return false // Unable to subscribe
	}

	logging.LogTarget("service", "subscribe", ssid)
	return true
}

// Occurs when a peer has unsubscribed.
func (s *Service) onUnsubscribe(ssid message.Ssid, sub message.Subscriber) (ok bool) {
	subscribers := s.subscriptions.Lookup(ssid)
	if ok = subscribers.Contains(sub); ok {
		s.subscriptions.Unsubscribe(ssid, sub)

		logging.LogTarget("service", "unsubscribe", ssid)
	}
	return
}

// Occurs when a message is received from a peer.
func (s *Service) onPeerMessage(m *message.Message) {
	// Get the contract
	contract, contractFound := s.contracts.Get(m.Ssid.Contract())

	// Iterate through all subscribers and send them the message
	for _, subscriber := range s.subscriptions.Lookup(m.Ssid) {
		if subscriber.Type() == message.SubscriberDirect {

			// Send to the local subscriber
			subscriber.Send(m)

			// Write the egress stats
			if contractFound {
				contract.Stats().AddEgress(int64(len(m.Payload)))
			}
		}
	}
}

// Query sends out a query to all the peers.
func (s *Service) Query(query string, payload []byte) (message.Awaiter, error) {
	if s.querier != nil {
		return s.querier.Query(query, payload)
	}

	return nil, errors.New("Query manager was not setup")
}

// Publish publishes a message to everyone and returns the number of outgoing bytes written.
func (s *Service) publish(ssid message.Ssid, channel, payload []byte) (n int64) {
	size := int64(len(payload))
	for _, subscriber := range s.subscriptions.Lookup(ssid) {
		subscriber.Send(&message.Message{
			Ssid:    ssid,
			Channel: channel,
			Payload: payload,
		})

		// Increment the egress size only for direct subscribers
		if subscriber.Type() == message.SubscriberDirect {
			n += size
		}
	}

	return
}

// SelfPublish publishes a message to itself.
func (s *Service) selfPublish(channelName string, payload []byte) {
	channel := security.ParseChannel([]byte("emitter/" + channelName))
	if channel.ChannelType == security.ChannelStatic {
		s.publish(message.NewSsid(s.License.Contract, channel), channel.Channel, payload)
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

	// Gracefully leave the cluster and shutdown the listener.
	if s.cluster != nil {
		_ = s.cluster.Close()
	}

	// Notify we're closed
	close(s.Closing)
}
