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
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http/pprof"

	"github.com/emitter-io/emitter/broker/cluster"
	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/emitter/network/listener"
	"github.com/emitter-io/emitter/network/tcp"
	"github.com/emitter-io/emitter/network/websocket"
	"github.com/emitter-io/emitter/security"
	"github.com/emitter-io/emitter/utils"
)

// Service represents the main structure.
type Service struct {
	Closing       chan bool                 // The channel for closing signal.
	Cipher        *security.Cipher          // The cipher to use for decoding and encoding keys.
	License       *security.License         // The licence for this emitter server.
	Config        *config.Config            // The configuration for the service.
	Contracts     security.ContractProvider // The contract provider for the service.
	subscriptions *SubscriptionTrie         // The subscription matching trie.
	subcounters   *SubscriptionCounters     // The subscription counters.
	http          *http.Server              // The underlying HTTP server.
	tcp           *tcp.Server               // The underlying TCP server.
	cluster       *cluster.Swarm            // The gossip-based cluster mechanism.
	startTime     time.Time                 // The start time of the service.
}

// NewService creates a new service.
func NewService(cfg *config.Config) (s *Service, err error) {
	s = &Service{
		Closing:       make(chan bool),
		Config:        cfg,
		subscriptions: NewSubscriptionTrie(),
		subcounters:   NewSubscriptionCounters(),
		http:          new(http.Server),
		tcp:           new(tcp.Server),
	}

	// Create a new HTTP request multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/keygen", s.onHTTPKeyGen)
	mux.HandleFunc("/debug/pprof/", pprof.Index)          // TODO: use config flag to enable/disable this
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline) // TODO: use config flag to enable/disable this
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile) // TODO: use config flag to enable/disable this
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)   // TODO: use config flag to enable/disable this
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)     // TODO: use config flag to enable/disable this
	mux.HandleFunc("/", s.onRequest)

	// Attach handlers
	s.http.Handler = mux
	s.tcp.Handler = s.onAcceptConn

	// Parse the license
	logging.LogAction("service", "external address is "+address.External().String())
	logging.LogAction("service", "reading the license...")
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
		s.cluster.OnSubscribe = s.onSubscribe
		s.cluster.OnUnsubscribe = s.onUnsubscribe
		s.cluster.OnMessage = s.onPeerMessage
	}

	return s, nil
}

// LocalName returns the local node name.
func (s *Service) LocalName() string {
	if s.cluster == nil {
		return address.Hardware().String()
	}

	return s.cluster.LocalName()
}

// Listen starts the service.
func (s *Service) Listen() (err error) {
	defer s.Close()
	s.hookSignals()

	// Create the cluster if required
	if s.cluster != nil {
		if s.cluster.Listen(); err != nil {
			panic(err)
		}

		// Join our seed
		s.Join(s.Config.Cluster.Seed)
	}

	// Setup the HTTP server
	logging.LogAction("service", "starting the listener...")
	l, err := listener.New(s.Config.TCPPort)
	if err != nil {
		panic(err)
	}

	l.ServeAsync(listener.MatchHTTP(), s.http.Serve)
	l.ServeAsync(listener.MatchAny(), s.tcp.Serve)

	// Set the start time and report status
	s.startTime = time.Now().UTC()
	utils.Repeat(s.reportStatus, 100*time.Millisecond, s.Closing)
	logging.LogAction("service", "service started")

	// Serve the listener
	if l.Serve(); err != nil {
		logging.LogError("service", "starting the listener", err)
	}
	return nil
}

// Join attempts to join a set of existing peers.
func (s *Service) Join(peers ...string) []error {
	return s.cluster.Join(peers...)
}

// NotifySubscribe notifies the swarm when a subscription occurs.
func (s *Service) notifySubscribe(conn *Conn, ssid []uint32) {
	if s.cluster != nil {
		s.cluster.NotifySubscribe(conn.id, ssid)
	}
}

// NotifyUnsubscribe notifies the swarm when an unsubscription occurs.
func (s *Service) notifyUnsubscribe(conn *Conn, ssid []uint32) {
	if s.cluster != nil {
		s.cluster.NotifyUnsubscribe(conn.id, ssid)
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
func (s *Service) onSubscribe(peer *cluster.Peer, event cluster.SubscriptionEvent) {
	fmt.Printf("%v subscribed to ssid: %v\n", event.Peer, event.Ssid)
	//s.subscriptions.Subscribe(event.Ssid, event.Channel, peer) TODO ! Figure out if we can get rid of channel string here
	s.subscriptions.Subscribe(event.Ssid, "TODO", peer)
}

// Occurs when a peer has unsubscribed.
func (s *Service) onUnsubscribe(peer *cluster.Peer, event cluster.SubscriptionEvent) {
	fmt.Printf("%v unsubscribed from ssid: %v\n", event.Peer, event.Ssid)

	s.subscriptions.Unsubscribe(event.Ssid, peer)
}

// Occurs when a message is received from a peer.
func (s *Service) onPeerMessage(m *cluster.Message) {
	fmt.Printf("message from peer on '%v' \n", string(m.Channel))

	// Get the contract
	ssid := Ssid(m.Ssid)
	contract := s.Contracts.Get(ssid.Contract())

	// Iterate through all subscribers and send them the message
	for _, subscriber := range s.subscriptions.Lookup(ssid) {
		if _, local := subscriber.(*Conn); local {

			// Send to the local subscriber
			subscriber.Send(m.Ssid, m.Channel, m.Payload)

			// Write the egress stats
			if contract != nil {
				contract.Stats().AddEgress(int64(len(m.Payload)))
			}
		}
	}
}

// SelfPublish publishes a message to itself.
func (s *Service) selfPublish(channelName string, payload []byte) {

	// Parse the channel and make an SSID we can use
	channel := security.ParseChannel([]byte("emitter/" + channelName))
	if channel.ChannelType == security.ChannelStatic {
		ssid := NewSsid(s.License.Contract, channel)

		// Iterate through all subscribers and send them the message
		subs := s.subscriptions.Lookup(ssid)
		//println("subscriber found: " + strconv.Itoa(len(subs)))
		for _, subscriber := range subs {
			subscriber.Send(ssid, channel.Channel, payload)
		}
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
