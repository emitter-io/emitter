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
	"strings"
	"syscall"
	"time"

	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/emitter/network/listener"
	"github.com/emitter-io/emitter/network/tcp"
	"github.com/emitter-io/emitter/network/websocket"
	"github.com/emitter-io/emitter/perf"
	"github.com/emitter-io/emitter/security"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
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
	cluster          *serf.Serf                // The gossip-based cluster mechanism.
	events           chan serf.Event           // The channel for receiving gossip events.
}

// NewService creates a new service.
func NewService(cfg *config.Config) (s *Service, err error) {
	s = &Service{
		Closing:       make(chan bool),
		Counters:      perf.NewCounters(),
		Config:        cfg,
		subscriptions: NewSubscriptionTrie(),
		events:        make(chan serf.Event),
		http:          new(http.Server),
		tcp:           new(tcp.Server),
	}

	// Attach handlers
	s.tcp.Handler = s.onAccept
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

	go func() {
		select {
		case e := <-s.events:
			println("Event: " + e.String())
		}
	}()

	// Create the cluster
	if s.cluster, err = serf.Create(s.clusterConfig(cfg)); err != nil {
		return nil, err
	}

	// Hook up signal handling
	s.hookSignals()
	return s, nil
}

// Creates a configuration for the cluster
func (s *Service) clusterConfig(cfg *config.Config) *serf.Config {
	c := serf.DefaultConfig()
	c.RejoinAfterLeave = true
	c.NodeName = address.Fingerprint() //fmt.Sprintf("%s:%d", address.External().String(), cfg.Cluster.Port) // TODO: fix this
	c.EventCh = s.events
	c.SnapshotPath = "cluster.log"
	c.MemberlistConfig = memberlist.DefaultWANConfig()
	c.MemberlistConfig.BindPort = cfg.Cluster.Port
	c.MemberlistConfig.AdvertisePort = cfg.Cluster.Port
	c.MemberlistConfig.SecretKey = cfg.Cluster.Key()

	// Set the node name
	c.NodeName = cfg.Cluster.NodeName
	if c.NodeName == "" {
		c.NodeName = fmt.Sprintf("%s%d", address.Fingerprint(), cfg.Cluster.Port)
	}

	// Use the public IP address if necessary
	if cfg.Cluster.Broadcast == "public" {
		c.MemberlistConfig.AdvertiseAddr = address.External().String()
	}

	return c
}

// OnSignal starts the signal processing and makes su
func (s *Service) hookSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range c {
			s.OnSignal(sig)
		}
	}()
}

// Listen starts the service.
func (s *Service) Listen() {
	defer logging.Flush()

	// Join our seed
	s.Join(s.Config.Cluster.Seed)

	go func() {
		for {
			members := []string{}
			for _, m := range s.cluster.Members() {
				members = append(members, fmt.Sprintf("%s (%s)", m.Name, m.Status.String()))
			}

			println(strings.Join(members, ", "))
			time.Sleep(1000 * time.Millisecond)
		}
	}()

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
}

// Join attempts to join a set of existing peers.
func (s *Service) Join(peers ...string) error {
	_, err := s.cluster.Join(peers, true)
	return err
}

// Broadcast is used to broadcast a custom user event with a given name and
// payload. The events must be fairly small, and if the  size limit is exceeded
// and error will be returned. If coalesce is enabled, nodes are allowed to
// coalesce this event.
func (s *Service) Broadcast(name string, payload []byte, coalesce bool) error {
	return s.cluster.UserEvent(name, payload, coalesce)
}

// Occurs when a new connection is accepted.
func (s *Service) onAccept(t net.Conn) {
	conn := s.newConn(t)
	go conn.Process()
}

// Occurs when a new HTTP request is received.
func (s *Service) onRequest(w http.ResponseWriter, r *http.Request) {
	if ws, ok := websocket.TryUpgrade(w, r); ok {
		s.onAccept(ws)
		return
	}
}

// OnSignal will be called when a OS-level signal is received.
func (s *Service) OnSignal(sig os.Signal) {
	switch sig {
	case syscall.SIGTERM:
		fallthrough
	case syscall.SIGINT:
		logging.LogAction("service", fmt.Sprintf("received signal %s, exiting...", sig.String()))
		_ = logging.Flush()

		// Gracefully leave the cluster and shutdown the listener.
		if s.cluster != nil {
			_ = s.cluster.Leave()
			_ = s.cluster.Shutdown()
		}

		// Notify we're closed
		close(s.Closing)
		os.Exit(0)
	}
}
