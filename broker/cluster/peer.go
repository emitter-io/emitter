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
	"sync"
	"time"

	"github.com/emitter-io/emitter/encoding"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/utils"
	"github.com/golang/snappy"
	"github.com/weaveworks/mesh"
)

// PeerSet maintains a map of peers
type peerset struct {
	sync.Mutex
	sender  mesh.Gossip             //  The gossip interface to use for sending.
	peers   *mesh.Peers             // The memberlist discovery mechanism.
	members map[mesh.PeerName]*Peer // The map of members in the peer set.
}

// newPeerSet creates a new peer set for the connection.
func newPeerSet(sender mesh.Gossip, peers *mesh.Peers) *peerset {
	p := &peerset{
		sender:  sender,
		peers:   peers,
		members: make(map[mesh.PeerName]*Peer),
	}

	// Attach the GC callback so we know when a peer is dead.
	peers.OnGC(p.onPeerGC)
	return p
}

// Occurs when a peer is garbage collected.
func (s *peerset) onPeerGC(peer *mesh.Peer) {
	if p := s.Get(peer.Name); p != nil {
		logging.LogTarget("swarm", "peer garbage collected", peer)
		p.Close() // Close the peer on our end

		// We also need to remove the peer from our set, so
		// the next time a new peer can be created.
		s.Lock()
		delete(s.members, p.name)
		s.Unlock()
	}
}

// Get retrieves a peer.
func (s *peerset) Get(name mesh.PeerName) (p *Peer) {
	s.Lock() // TODO: This lock will be contended eventually, need to replace this map by sync.Map
	defer s.Unlock()

	// Get the peer
	if p, ok := s.members[name]; ok {
		return p
	}

	// Create new peer
	p = s.newPeer(name)
	s.members[name] = p
	return p
}

// ------------------------------------------------------------------------------------

// Peer represents a remote peer.
type Peer struct {
	sync.Mutex
	sender  mesh.Gossip   // The gossip interface to use for sending.
	name    mesh.PeerName // The peer name for communicating.
	frame   MessageFrame  // The current message frame.
	closing chan bool     // The closing channel for the peer.
}

// NewPeer creates a new peer for the connection.
func (s *peerset) newPeer(name mesh.PeerName) *Peer {
	peer := &Peer{
		sender:  s.sender,
		name:    name,
		frame:   make(MessageFrame, 0, 64),
		closing: make(chan bool),
	}

	// Spawn the send queue processor
	utils.Repeat(peer.processSendQueue, 5*time.Millisecond, peer.closing)
	return peer
}

// Close termintes the peer and stops everything associated with this peer.
func (p *Peer) Close() {
	close(p.closing)
}

// Send forwards the message to the remote server.
func (p *Peer) Send(ssid []uint32, channel []byte, payload []byte) error {
	p.Lock()
	defer p.Unlock()

	// Send simply appends the message to a frame
	p.frame = append(p.frame, &Message{Ssid: ssid, Channel: channel, Payload: payload})
	return nil
}

// processSendQueue flushes the current frame to the remote server
func (p *Peer) processSendQueue() {
	if len(p.frame) > 0 {

		// Compress in-memory. TODO: Optimize the shit out of that, we don't really need to use binc
		buffer := bytes.NewBuffer(nil)
		snappy := snappy.NewBufferedWriter(buffer)
		writer := encoding.NewEncoder(snappy)

		// Encode the current frame
		p.Lock()
		err := writer.Encode(p.frame)
		p.frame = p.frame[:0]
		p.Unlock()

		// Something went wrong during the encoding
		if err != nil {
			logging.LogError("peer", "encoding frame", err)
		}

		// Send the frame directly to the peer.
		if err := snappy.Close(); err != nil {
			logging.LogError("peer", "encoding frame", err)
		}
		p.sender.GossipUnicast(p.name, buffer.Bytes())
	}
}
