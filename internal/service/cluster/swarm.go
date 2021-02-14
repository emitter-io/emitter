/**********************************************************************************
* Copyright (c) 2009-2020 Misakai Ltd.
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

package cluster

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/emitter-io/address"
	"github.com/emitter-io/emitter/internal/async"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/weaveworks/mesh"
)

// Swarm represents a gossiper.
type Swarm struct {
	sync.Mutex
	name    mesh.PeerName         // The name of ourselves.
	actions chan func()           // The action queue for the peer.
	cancel  context.CancelFunc    // The cancellation function.
	config  *config.ClusterConfig // The configuration for the cluster.
	state   *event.State          // The state to synchronise.
	router  *mesh.Router          // The mesh router.
	gossip  mesh.Gossip           // The gossip protocol.
	members *memberlist           // The memberlist of peers.

	OnSubscribe   func(message.Subscriber, *event.Subscription) bool // Delegate to invoke when the subscription event is received.
	OnUnsubscribe func(message.Subscriber, *event.Subscription) bool // Delegate to invoke when the unsubscription event is received.
	OnDisconnect  func(message.Subscriber, *event.Connection) bool   // Delegate to invoke when the client is disconnected.
	OnMessage     func(*message.Message)                             // Delegate to invoke when a new message is received.
}

// Swarm implements mesh.Gossiper.
var _ mesh.Gossiper = &Swarm{}

// NewSwarm creates a new swarm messaging layer.
func NewSwarm(cfg *config.ClusterConfig) *Swarm {
	name := getLocalPeerName(cfg)
	if d, err := os.UserCacheDir(); cfg.Directory == "" && err == nil {
		cfg.Directory = path.Join(d, fmt.Sprintf("emitter/%x", int(name)))
	}

	os.MkdirAll(cfg.Directory, os.ModePerm)
	swarm := &Swarm{
		name:    name,
		actions: make(chan func()),
		config:  cfg,
		state:   event.NewState(cfg.Directory),
	}

	// Get the cluster binding address
	listenAddr, err := address.Parse(cfg.ListenAddr, 4000)
	if err != nil {
		panic(err)
	}

	// Get the advertised address
	advertiseAddr, err := address.Parse(cfg.AdvertiseAddr, 4000)
	if err != nil {
		panic(err)
	}

	// Create a new router
	interval := 30 * time.Second
	router, err := mesh.NewRouter(mesh.Config{
		Host:               listenAddr.IP.String(),
		Port:               listenAddr.Port,
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		Password:           []byte(cfg.Passphrase),
		ConnLimit:          128,
		PeerDiscovery:      true,
		TrustedSubnets:     []*net.IPNet{},
		GossipInterval:     &interval,
	}, swarm.name, advertiseAddr.String(), mesh.NullOverlay{}, swarm)
	if err != nil {
		panic(err)
	}

	// Handle when peer is removed
	router.Peers.OnGC(func(peer *mesh.Peer) {
		swarm.onPeerOffline(peer.Name)
	})

	// Create a new gossip layer
	gossip, err := router.NewGossip("swarm", swarm)
	if err != nil {
		panic(err)
	}

	//Store the gossip and the router
	swarm.gossip = gossip
	swarm.router = router
	swarm.members = newMemberlist(swarm.newPeer)
	return swarm
}

// Printf logs the error
func (s *Swarm) Printf(format string, args ...interface{}) {
	message := fmt.Errorf(format, args...)
	if strings.Contains(message.Error(), "error") {
		logging.LogError("swarm", "gossip", message)
	}
}

// findPeer retrieves a peer.
func (s *Swarm) findPeer(name mesh.PeerName) *Peer {
	peer, added := s.members.GetOrAdd(name)
	if added {
		s.onPeerOnline(peer)
	}

	return peer
}

// onPeerOnline occurs when a new peer is created.
func (s *Swarm) onPeerOnline(peer *Peer) {
	logging.LogTarget("swarm", "peer created", peer.name)
	s.state.SubscriptionsOf(peer.name, func(ev *event.Subscription) {
		s.OnSubscribe(peer, ev)
	})
}

// Occurs when a peer is garbage collected.
func (s *Swarm) onPeerOffline(name mesh.PeerName) {
	if peer, deleted := s.members.Remove(name); deleted {
		logging.LogTarget("swarm", "unreachable peer removed", peer.name)
		peer.Close() // Close the peer on our end

		// Range over all of the subscriptions we have
		dead := &deadPeer{name: name}
		s.state.SubscriptionsOf(name, func(ev *event.Subscription) {
			s.OnUnsubscribe(dead, ev) // Notify locally that the subscription is gone
			s.state.Del(ev)           // Remove the state from ourselves
		})

		// If we're a fallback server, issue last will events
		if fallback, ok := s.members.Fallback(name); ok && s.name == fallback.name {
			s.state.ConnectionsOf(name, func(ev *event.Connection) {
				s.OnDisconnect(dead, ev)
			})
		}
	}
}

// SendTo sends a message to a peer.
func (s *Swarm) SendTo(name mesh.PeerName, msg *message.Message) error {
	peer := s.findPeer(name)
	if !peer.IsActive() {
		return errors.New("swarm: unable to reply to a request, peer is not active")
	}

	return peer.Send(msg)
}

// ID returns the local node ID.
func (s *Swarm) ID() uint64 {
	return uint64(s.name)
}

// Listen creates the listener and serves the cluster.
func (s *Swarm) Listen(ctx context.Context) {

	// Every few seconds, attempt to reinforce our cluster structure by
	// initiating connections with all of our peers.
	s.cancel = async.Repeat(ctx, 5*time.Second, s.update)

	// Start the router
	s.router.Start()
}

// update attempt to update our cluster structure by initiating connections
// with all of our peers. This is is called periodically.
func (s *Swarm) update() {
	desc := s.router.Peers.Descriptions()
	for _, peer := range desc {
		if !peer.Self {

			// Mark the peer as active, so even if there's no messages being exchanged
			// we still keep the peer, since we know that the peer is live.
			if exists := s.router.Peers.Fetch(peer.Name); exists != nil {
				s.members.Touch(peer.Name)
			}

			// reinforce structure
			if peer.NumConnections < (len(desc) - 1) {
				s.Join(peer.NickName)
			}
		}
	}
}

// Join attempts to join a set of existing peers.
func (s *Swarm) Join(peers ...string) (errs []error) {
	// Resolve the host-names of the peers provided
	var addrs []string
	for _, h := range peers {

		// Check first if this is a domain name and add its addresses
		if ips, err := net.LookupHost(h); err == nil {
			addrs = append(addrs, ips...)
			continue
		}

		// It's not a host name, parse the address
		addr, err := address.Parse(h, 80)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// Successfullyt parsed the address
		addrs = append(addrs, addr.String())
	}

	for _, a := range addrs {
		logging.LogTarget("swarm", "joining", a)
	}

	// Use all the available addresses to initiate the connections
	if s.router != nil {
		errs = s.router.ConnectionMaker.InitiateConnections(addrs, false)
	}
	return
}

// Merge merges the incoming state and returns a delta
func (s *Swarm) merge(buf []byte) (mesh.GossipData, error) {

	// Decode the state we just received
	other, err := event.DecodeState(buf)
	if err != nil {
		return nil, err
	}

	// Merge and get the delta
	delta := s.state.Merge(other)
	other.Subscriptions(func(ev *event.Subscription, v event.Value) {
		if ev.Peer == uint64(s.router.Ourself.Name) {
			return // Skip ourselves
		}

		// Find the active peer for this subscription event
		key := ev.Key()
		peer := s.findPeer(mesh.PeerName(ev.Peer))

		// If the subscription is added, notify (TODO: use channels)
		if v.IsAdded() && peer.onSubscribe(key, ev.Ssid) && peer.IsActive() {
			s.OnSubscribe(peer, ev)
		}

		// If the subscription is removed, notify (TODO: use channels)
		if v.IsRemoved() && peer.onUnsubscribe(key, ev.Ssid) && peer.IsActive() {
			s.OnUnsubscribe(peer, ev)
		}
	})

	return delta, nil
}

// NumPeers returns the number of connected peers.
func (s *Swarm) NumPeers() int {
	if s == nil || s.router == nil {
		return 0
	}

	for _, peer := range s.router.Peers.Descriptions() {
		if peer.Self {
			return peer.NumConnections
		}
	}
	return 0
}

// Gossip returns the state of everything we know; gets called periodically.
func (s *Swarm) Gossip() (complete mesh.GossipData) {
	return s.state
}

// OnGossip merges received data into state and returns "everything new I've just
// learnt", or nil if nothing in the received data was new.
func (s *Swarm) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	if len(buf) <= 1 {
		return nil, nil
	}

	if delta, err = s.merge(buf); err != nil {
		logging.LogError("merge", "merging", err)
	}
	return
}

// OnGossipBroadcast merges received data into state and returns a representation
// of the received data (typically a delta) for further propagation.
func (s *Swarm) OnGossipBroadcast(src mesh.PeerName, buf []byte) (delta mesh.GossipData, err error) {
	if src == s.name {
		logging.LogAction("merge", "got our own broadcast")
		return
	}

	if delta, err = s.merge(buf); err != nil {
		logging.LogError("merge", "merging", err)
	}
	return
}

// OnGossipUnicast occurs when the gossip unicast is received. In emitter this is
// used only to forward message frames around.
func (s *Swarm) OnGossipUnicast(src mesh.PeerName, buf []byte) (err error) {

	// Decode an incoming message frame
	frame, err := message.DecodeFrame(buf)
	if err != nil {
		logging.LogError("swarm", "decode frame", err)
		return err
	}

	// Go through each message in the decoded frame
	for i := range frame {
		s.OnMessage(&frame[i])
	}

	return nil
}

// Notify notifies the swarm when an event is on/off.
func (s *Swarm) Notify(ev event.Event, enabled bool) {
	op := event.NewState("")
	if enabled {
		s.state.Add(ev)
		op.Add(ev)
	} else {
		s.state.Del(ev)
		op.Del(ev)
	}

	// Broadcasting just this operation
	s.gossip.GossipBroadcast(op)
}

// Contains checks whether an event is currently triggered within the cluster.
func (s *Swarm) Contains(ev event.Event) bool {
	return s.state.Has(ev)
}

// Close terminates the connection.
func (s *Swarm) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	s.state.Close()
	return s.router.Stop()
}

// getLocalPeerName retrieves or generates a local node name.
func getLocalPeerName(cfg *config.ClusterConfig) mesh.PeerName {
	peerName := mesh.PeerName(address.GetHardware())
	if cfg.NodeName != "" {
		if name, err := mesh.PeerNameFromString(cfg.NodeName); err != nil {
			logging.LogError("swarm", "getting node name", err)
		} else {
			peerName = name
		}
	}

	return peerName
}
