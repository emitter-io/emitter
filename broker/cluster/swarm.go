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

package cluster

import (
	"bytes"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/encoding"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/emitter/security"
	"github.com/emitter-io/emitter/utils"
	"github.com/golang/snappy"
	"github.com/weaveworks/mesh"
)

// Swarm represents a gossiper.
type Swarm struct {
	sync.Mutex
	name    mesh.PeerName         // The name of ourselves.
	actions chan func()           // The action queue for the peer.
	closing chan bool             // The closing channel.
	config  *config.ClusterConfig // The configuration for the cluster.
	state   *subscriptionState    // The state to synchronise.
	router  *mesh.Router          // The mesh router.
	gossip  mesh.Gossip           // The gossip protocol.
	members *sync.Map             // The map of members in the peer set.

	OnSubscribe   func(subscription.Ssid, subscription.Subscriber) bool // Delegate to invoke when the subscription event is received.
	OnUnsubscribe func(subscription.Ssid, subscription.Subscriber) bool // Delegate to invoke when the subscription event is received.
	OnMessage     func(*Message)                                        // Delegate to invoke when a new message is received.
}

// Swarm implements mesh.Gossiper.
var _ mesh.Gossiper = &Swarm{}

// NewSwarm creates a new swarm messaging layer.
func NewSwarm(cfg *config.ClusterConfig, closing chan bool) *Swarm {
	swarm := &Swarm{
		name:    getLocalPeerName(cfg),
		actions: make(chan func()),
		closing: closing,
		config:  cfg,
		state:   newSubscriptionState(),
		members: new(sync.Map),
	}

	// Get the cluster binding address
	clusterAddr, err := parseAddr(cfg.ClusterAddr)
	if err != nil {
		panic(err)
	}

	// Get the advertised address
	advertiseAddr, err := parseAddr(cfg.AdvertiseAddr)
	if err != nil {
		panic(err)
	}

	// Create a new router
	router, err := mesh.NewRouter(mesh.Config{
		Host:               clusterAddr.IP.String(),
		Port:               clusterAddr.Port,
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		Password:           []byte(cfg.ClusterKey),
		ConnLimit:          128,
		PeerDiscovery:      true,
		TrustedSubnets:     []*net.IPNet{},
	}, swarm.name, advertiseAddr.String(), mesh.NullOverlay{}, logging.Discard)
	if err != nil {
		panic(err)
	}

	// Create a new gossip layer
	gossip, err := router.NewGossip("swarm", swarm)
	if err != nil {
		panic(err)
	}

	//Store the gossip and the router
	swarm.gossip = gossip
	swarm.router = router
	return swarm
}

// Occurs when a peer is garbage collected.
func (s *Swarm) onPeerOffline(name mesh.PeerName) {
	if v, ok := s.members.Load(name); ok {
		peer := v.(*Peer)
		logging.LogTarget("swarm", "peer removed", peer.name)
		peer.Close() // Close the peer on our end

		// We also need to remove the peer from our set, so next time a new peer can be created.
		s.members.Delete(peer.name)

		// Unsubscribe from all active subscriptions
		for _, ssid := range peer.subs {
			s.OnUnsubscribe(ssid, peer)
		}
	}
}

// Get retrieves a peer.
func (s *Swarm) findPeer(name mesh.PeerName) *Peer {
	if p, ok := s.members.Load(name); ok {
		return p.(*Peer)
	}

	// Create new peer and store it
	peer := s.newPeer(name)
	v, ok := s.members.LoadOrStore(name, peer)
	if !ok {
		logging.LogTarget("swarm", "peer created", peer.name)
	}
	return v.(*Peer)
}

// LocalName returns the local node name.
func (s *Swarm) LocalName() string {
	return s.name.String()
}

// Listen creates the listener and serves the cluster.
func (s *Swarm) Listen() {

	// Every few seconds, attempt to reinforce our cluster structure by
	// initiating connections with all of our peers.
	utils.Repeat(s.update, 5*time.Second, s.closing)

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
			s.findPeer(peer.Name).touch()

			// reinforce structure
			if peer.NumConnections < (len(desc) - 1) {
				s.Join(peer.NickName)
			}
		}
	}

	// Mark a peer as offline
	s.members.Range(func(k, v interface{}) bool {
		if p, ok := v.(*Peer); ok && !p.IsActive() {
			s.onPeerOffline(p.name)
		}
		return true
	})
}

// Join attempts to join a set of existing peers.
func (s *Swarm) Join(peers ...string) []error {
	return s.router.ConnectionMaker.InitiateConnections(peers, false)
}

// Merge merges the incoming state and returns a delta
func (s *Swarm) merge(buf []byte) (mesh.GossipData, error) {

	// Decode the state we just received
	other, err := decodeSubscriptionState(buf)
	if err != nil {
		return nil, err
	}

	// Merge and get the delta
	delta := s.state.Merge(other)
	for k, v := range other.All() {

		// Decode the event
		ev, err := decodeSubscriptionEvent(k.(string))
		if err != nil {
			return nil, err
		}

		// Get the peer to use
		peer := s.findPeer(ev.Peer)

		// If the subscription is added, notify (TODO: use channels)
		if v.IsAdded() {
			peer.onSubscribe(k.(string), ev.Ssid)
			s.OnSubscribe(ev.Ssid, peer)
		}

		// If the subscription is removed, notify (TODO: use channels)
		if v.IsRemoved() {
			peer.onUnsubscribe(k.(string), ev.Ssid)
			s.OnUnsubscribe(ev.Ssid, peer)
		}
	}

	return delta, nil
}

// NumPeers returns the number of connected peers.
func (s *Swarm) NumPeers() int {
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
	if delta, err = s.merge(buf); err != nil {
		logging.LogError("merge", "merging", err)
	}
	return
}

// OnGossipUnicast occurs when the gossip unicast is received. In emitter this is
// used only to forward message frames around.
func (s *Swarm) OnGossipUnicast(src mesh.PeerName, buf []byte) error {

	// Make a reader and a decoder for the frame
	snappy := snappy.NewReader(bytes.NewReader(buf))
	reader := encoding.NewDecoder(snappy)

	// Decode an incoming message frame
	frame, err := decodeMessageFrame(reader)
	if err != nil {
		logging.LogError("peer", "decode frame", err)
		return err
	}

	// Go through each message in the decoded frame
	for _, m := range frame {
		s.OnMessage(m)
	}

	return nil
}

// NotifySubscribe notifies the swarm when a subscription occurs.
func (s *Swarm) NotifySubscribe(conn security.ID, ssid []uint32) {
	event := SubscriptionEvent{
		Peer: s.name,
		Conn: conn,
		Ssid: ssid,
	}

	// Add to our global state
	s.state.Add(event.Encode())

	// Create a delta for broadcasting just this operation
	op := newSubscriptionState()
	op.Add(event.Encode())
	s.gossip.GossipBroadcast(op)
}

// NotifyUnsubscribe notifies the swarm when an unsubscription occurs.
func (s *Swarm) NotifyUnsubscribe(conn security.ID, ssid []uint32) {
	event := SubscriptionEvent{
		Peer: s.name,
		Conn: conn,
		Ssid: ssid,
	}

	// Remove from our global state
	s.state.Remove(event.Encode())

	// Create a delta for broadcasting just this operation
	op := newSubscriptionState()
	op.Remove(event.Encode())
	s.gossip.GossipBroadcast(op)
}

// Close terminates the connection.
func (s *Swarm) Close() error {
	return s.router.Stop()
}

// parseAddr parses a TCP address.
func parseAddr(text string) (*net.TCPAddr, error) {
	if text[0] == ':' {
		text = "0.0.0.0" + text
	}

	addr := strings.Replace(text, "public", address.External().String(), 1)
	return net.ResolveTCPAddr("tcp", addr)
}

// getLocalPeerName retrieves or generates a local node name.
func getLocalPeerName(cfg *config.ClusterConfig) mesh.PeerName {
	peerName := mesh.PeerName(address.Hardware())
	if cfg.NodeName != "" {
		if name, err := mesh.PeerNameFromString(cfg.NodeName); err != nil {
			logging.LogError("swarm", "getting node name", err)
		} else {
			peerName = name
		}
	}

	return peerName
}
