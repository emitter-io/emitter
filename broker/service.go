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

	"github.com/emitter-io/emitter/broker/cluster"
	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/emitter/network/listener"
	"github.com/emitter-io/emitter/network/tcp"
	"github.com/emitter-io/emitter/network/websocket"
	"github.com/emitter-io/emitter/security"
)

// Service represents the main structure.
type Service struct {
	Closing          chan bool                 // The channel for closing signal.
	Cipher           *security.Cipher          // The cipher to use for decoding and encoding keys.
	License          *security.License         // The licence for this emitter server.
	Config           *config.Config            // The configuration for the service.
	ContractProvider security.ContractProvider // The contract provider for the service.
	subscriptions    *SubscriptionTrie         // The subscription matching trie.
	subcounters      *SubscriptionCounters     // The subscription counters.
	http             *http.Server              // The underlying HTTP server.
	tcp              *tcp.Server               // The underlying TCP server.
	cluster          *cluster.Cluster          // The gossip-based cluster mechanism.
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
		if s.cluster, err = cluster.NewCluster(cfg.Cluster, s.Closing); err != nil {
			return nil, err
		}

		// Attach delegates
		s.cluster.OnSubscribe = s.onSubscribe
		s.cluster.OnUnsubscribe = s.onUnsubscribe
		s.cluster.OnMessage = s.onPeerMessage
		s.cluster.Subscriptions = s.subcounters.All
	}

	return s, nil
}

// LocalName returns the local node name.
func (s *Service) LocalName() string {
	if s.cluster == nil {
		return address.Fingerprint()
	}

	return s.cluster.LocalName()
}

// Listen starts the service.
func (s *Service) Listen() (err error) {
	defer s.Close()
	s.hookSignals()

	// Create the cluster if required
	if s.cluster != nil {
		if s.cluster.Listen(s.Config.Cluster.Route); err != nil {
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

	// Serve the listener
	if l.Serve(); err != nil {
		logging.LogError("service", "starting the listener", err)
	}

	return nil
}

// Join attempts to join a set of existing peers.
func (s *Service) Join(peers ...string) error {
	return s.cluster.Join(peers...)
}

// Broadcast is used to broadcast a custom user event with a given name and
// payload. The events must be fairly small, and if the  size limit is exceeded
// and error will be returned. If coalesce is enabled, nodes are allowed to
// coalesce this event.
func (s *Service) Broadcast(name string, message interface{}) error {
	if s.cluster == nil {
		return nil
	}

	return s.cluster.Broadcast(name, message)
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
	fmt.Printf("%v subscribed to ssid: %v\n", event.Node, event.Ssid)
	s.subscriptions.Subscribe(event.Ssid, event.Channel, peer)
}

// Occurs when a peer has unsubscribed.
func (s *Service) onUnsubscribe(peer *cluster.Peer, event cluster.SubscriptionEvent) {
	fmt.Printf("%v unsubscribed from ssid: %v\n", event.Node, event.Ssid)

	s.subscriptions.Unsubscribe(event.Ssid, peer)
}

// Occurs when a message is received from a peer.
func (s *Service) onPeerMessage(m *cluster.Message) {
	fmt.Printf("message from peer on '%v' \n", string(m.Channel))

	// Iterate through all subscribers and send them the message
	for _, subscriber := range s.subscriptions.Lookup(Ssid(m.Ssid)) {
		if _, local := subscriber.(*Conn); local {
			subscriber.Send(m.Ssid, m.Channel, m.Payload)
		}
	}
}

// Occurs when a query is received.
func (s *Service) onQuery(query cluster.Query) {
	fmt.Printf("query: %v\n", query)
}

// Occurs when a query response is received from a node.
func (s *Service) onQueryResponse(resp cluster.QueryResponse) {
	fmt.Printf("query response: %v\n", resp.Payload)
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
	_ = logging.Flush()

	// Gracefully leave the cluster and shutdown the listener.
	if s.cluster != nil {
		_ = s.cluster.Close()
	}

	// Notify we're closed
	close(s.Closing)
}
