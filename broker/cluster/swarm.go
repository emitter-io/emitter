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
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
	"time"

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

	OnSubscribe   func(*Peer, SubscriptionEvent) // Delegate to invoke when the subscription event is received.
	OnUnsubscribe func(*Peer, SubscriptionEvent) // Delegate to invoke when the subscription event is received.
	OnMessage     func(*Message)                 // Delegate to invoke when a new message is received.
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
	}, swarm.name, advertiseAddr.String(), mesh.NullOverlay{}, log.New(ioutil.Discard, "", 0))
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

// LocalName returns the local node name.
func (s *Swarm) LocalName() string {
	return s.name.String()
}

// Listen creates the listener and serves the cluster.
func (s *Swarm) Listen() (err error) {

	// Every few seconds, attempt to reinforce our cluster structure by
	// initiating connections with all of our peers.
	utils.Repeat(s.reinforce, 5*time.Second, s.closing)

	// Start processing action queue
	go s.loop()

	// Start the router
	s.router.Start()
	return nil
}

// reinforce attempt to reinforce our cluster structure by initiating connections
// with all of our peers. This is is called periodically.
func (s *Swarm) reinforce() {
	peers := []string{}
	for _, peer := range s.router.Peers.Descriptions() {
		if !peer.Self {
			peers = append(peers, peer.NickName)
		}
	}

	s.router.ConnectionMaker.InitiateConnections(peers, false)
}

// Join attempts to join a set of existing peers.
func (s *Swarm) Join(peers ...string) []error {
	return s.router.ConnectionMaker.InitiateConnections(peers, true)
}

// loop processes action queue
func (s *Swarm) loop() {
	for {
		select {
		case f := <-s.actions:
			f()

		case <-s.closing:
			_ = s.Close()
			return
		}
	}
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

		if v.IsAdded() {
			logging.LogTarget("swarm", "subscribe", ev)
			s.OnSubscribe(newPeer(s, ev.Peer), ev)
		}

		if v.IsRemoved() {
			logging.LogTarget("swarm", "unsubscribe", ev)
			s.OnUnsubscribe(newPeer(s, ev.Peer), ev)
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

	logging.LogAction("peer", "OnGossip(): "+fmt.Sprintf("%v bytes received", len(buf)))

	if delta, err = s.merge(buf); err != nil {
		logging.LogError("merge", "merging", err)
	}
	return
}

// OnGossipBroadcast merges received data into state and returns a representation
// of the received data (typically a delta) for further propagation.
func (s *Swarm) OnGossipBroadcast(src mesh.PeerName, buf []byte) (delta mesh.GossipData, err error) {
	logging.LogAction("peer", "OnGossipBroadcast(): "+fmt.Sprintf("%v bytes received", len(buf)))

	if delta, err = s.merge(buf); err != nil {
		logging.LogError("merge", "merging", err)
	}
	return
}

// OnGossipUnicast occurs when the gossip unicast is received. In emitter this is
// used only to forward message frames around.
func (s *Swarm) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	logging.LogAction("peer", "OnGossipUnicast()")

	// Make a reader and a decoder for the frame
	reader := snappy.NewReader(bytes.NewReader(buf))
	decoder := encoding.NewDecoder(reader)

	// Decode an incoming message frame
	frame, err := decodeMessageFrame(decoder)
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
