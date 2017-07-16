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
	"github.com/emitter-io/emitter/perf"
	"github.com/emitter-io/emitter/security"
)

// Service represents the main structure.
type Service struct {
	Closing          chan bool                 // The channel for closing signal.
	Counters         *perf.Counters            // The performance counters for this service.
	Cipher           *security.Cipher          // The cipher to use for decoding and encoding keys.
	License          *security.License         // The licence for this emitter server.
	Config           *config.Config            // The configuration for the service.
	ContractProvider security.ContractProvider // The contract provider for the service.
	subscriptions    *SubscriptionTrie         // The subscription matching trie.
	http             *http.Server              // The underlying HTTP server.
	tcp              *tcp.Server               // The underlying TCP server.
	cluster          *cluster.Cluster          // The gossip-based cluster mechanism.
}

// NewService creates a new service.
func NewService(cfg *config.Config) (s *Service, err error) {
	s = &Service{
		Closing:       make(chan bool),
		Counters:      perf.NewCounters(),
		Config:        cfg,
		subscriptions: NewSubscriptionTrie(),
		http:          new(http.Server),
		tcp:           new(tcp.Server),
	}

	// Attach handlers
	s.tcp.Handler = s.onAcceptConn
	http.HandleFunc("/", s.onRequest)

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
	}

	return s, nil
}

// LocalName returns the local node name.
func (s *Service) LocalName() string {
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
	}

	// Join our seed
	s.Join(s.Config.Cluster.Seed)

	/*go func() {
		for {
			members := []string{}
			for _, m := range s.cluster.Members() {
				members = append(members, fmt.Sprintf("%s (%s)", m.Name, m.Status.String()))
			}

			println(strings.Join(members, ", "))
			time.Sleep(1000 * time.Millisecond)
		}
	}()*/

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

// Occurs when a peer has a new subscription.
func (s *Service) onSubscribe(event *cluster.SubscriptionEvent) {

}

// Occurs when a peer has unsubscribed.
func (s *Service) onUnsubscribe(event *cluster.SubscriptionEvent) {

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
