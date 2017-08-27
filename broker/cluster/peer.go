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
	sync.RWMutex
	sender  mesh.Gossip             //  The gossip interface to use for sending.
	members map[mesh.PeerName]*Peer // The map of members in the peer set.
}

// newPeerSet creates a new peer set for the connection.
func newPeerSet(sender mesh.Gossip) *peerset {
	return &peerset{
		sender:  sender,
		members: make(map[mesh.PeerName]*Peer),
	}
}

// Get retrieves a peer.
func (s *peerset) Get(name mesh.PeerName) (p *Peer) {
	s.RLock() // TODO: This lock will be contended eventually, need to replace this map by sync.Map
	defer s.RUnlock()

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

/*
import (
	"bufio"
	"bytes"
	"net"
	"sync"

	"github.com/emitter-io/emitter/encoding"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/utils"
	"github.com/golang/snappy"
	"github.com/weaveworks/mesh"
)

const (
	readBufferSize = 1048576 // 1MB per peer
)

var logConnection = logging.AddLogger("[peer] connection %s (remote: %s)")

// Peer represents a peer broker.
type Peer struct {
	sync.Mutex
	send       mesh.Gossip   // The gossip protocol.
	actions    chan<- func() // The action queue for the peer.
	closing    chan bool     // The closing channel.
	frame      MessageFrame  // The current message frame.
	handshaken bool          // The flag for handshake.
	name       string        // The name of the peer.

	OnClosing   func(*Peer)                       // Handler which is invoked when the peer is closing is received.
	OnHandshake func(*Peer, HandshakeEvent) error // Handler which is invoked when a handshake is received.
	OnMessage   func(*Message)                    // Handler which is invoked when a new message is received.
}

// Peer implements mesh.Gossiper.
var _ mesh.Gossiper = &Peer{}

// NewPeer creates a new peer for the connection.
func newPeer() *Peer {
	logging.Log(logConnection, "opened", conn.RemoteAddr().String())
	actions := make(chan func())

	c := &Peer{
		send:    nil, // must .register() later
		actions: actions,
		closing: make(chan bool),
		socket:  conn,
		writer:  snappy.NewBufferedWriter(conn),
		reader:  snappy.NewReader(bufio.NewReaderSize(conn, readBufferSize)),
		frame:   make(MessageFrame, 0, 64),
	}

	// Start processing action queue
	go c.loop(actions)

	// Spawn the send queue processor as well
	//utils.Repeat(c.processSendQueue, 5*time.Millisecond, c.closing)
	return c
}

// loop processes action queue
func (c *Peer) loop(actions <-chan func()) {
	for {
		select {
		case f := <-actions:
			f()
		case <-c.closing:
			return
		}
	}
}

// Register the result of a mesh.Router.NewGossip.
func (c *Peer) Register(send mesh.Gossip) {
	c.send = send
}

// Send forwards the message to the remote server.
func (c *Peer) Send(ssid []uint32, channel []byte, payload []byte) error {
	c.Lock()
	defer c.Unlock()

	// Send simply appends the message to a frame
	c.frame = append(c.frame, &Message{Ssid: ssid, Channel: channel, Payload: payload})
	return nil
}

// processSendQueue flushes the current frame to the remote server
func (c *Peer) processSendQueue() {
	if len(c.frame) > 0 {
		encoder := encoding.NewEncoder(c.writer)

		// Encode the current frame
		c.Lock()
		err := encoder.Encode(c.frame)
		c.frame = c.frame[:0]
		c.Unlock()

		// Something went wrong during the encoding
		if err != nil {
			logging.LogError("peer", "encoding frame", err)
		}

		// Flush the writer
		c.send.GossipUnicast(c.)
		c.writer.Flush()
	}
}

// Handshake sends a handshake message to the peer.
func (c *Peer) Handshake(node string, subDelegate func() []Subscription) (err error) {
	c.Lock()
	defer c.Unlock()

	// Avoid sending the handshake recursively.
	if c.handshaken || node == "" {
		return
	}

	// Retrieve all existing subscriptions for the handshake
	var subs []Subscription
	if subDelegate != nil {
		subs = subDelegate()
	}

	// Send the handshake through
	c.handshaken = true
	err = encoding.EncodeTo(c.writer, &HandshakeEvent{
		Key:  "", // TODO add key
		Node: node,
		Subs: subs,
	})

	// Flush the buffered writer so we'd actually write through the socket
	if err == nil {
		err = c.writer.Flush()
	}

	return
}

// Process processes the messages.
func (c *Peer) Process() error {
	defer c.Close()
	decoder := encoding.NewDecoder(c.reader)

	for {
		// Decode an incoming message frame
		frame, err := decodeMessageFrame(decoder)
		if err != nil {
			logging.LogError("peer", "decode frame", err)
			return err
		}

		// Go through each message in the decoded frame
		for _, m := range frame {
			c.OnMessage(m)
		}
	}
}

// Gossip returns a copy of our complete state.
func (c *Peer) Gossip() (complete mesh.GossipData) {
	logging.LogAction("peer", "Gossip()")
	return nil
}

// OnGossip occurs when the peer receives the gossip message.
func (c *Peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	logging.LogAction("peer", "OnGossip()")
	return nil, nil
}

// OnGossipBroadcast occurs when the gossip broadcast is received.
func (c *Peer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	logging.LogAction("peer", "OnGossipBroadcast()")
	return received, nil
}

// OnGossipUnicast occurs when the gossip unicast is received.
func (c *Peer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
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
		c.OnMessage(m)
	}

	return nil
}

// Close terminates the connection.
func (c *Peer) Close() error {
	logging.Log(logConnection, "closed", c.socket.RemoteAddr().String())
	close(c.closing)

	// Close the peer.
	if c.OnClosing != nil {
		c.OnClosing(c)
	}

	// First we need to close the writer
	c.writer.Close()

	// Finally, close the underlying socket.
	return c.socket.Close()
}

// peerKey returns a peer key from a node name.
func peerKey(nodeName string) uint32 {
	return utils.GetHash([]byte(nodeName))
}
*/
