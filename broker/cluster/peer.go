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
	"bufio"
	"net"
	"sync"
	"time"

	"github.com/emitter-io/emitter/encoding"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/utils"
	"github.com/golang/snappy"
)

const (
	readBufferSize = 1048576 // 1MB per peer
)

var logConnection = logging.AddLogger("[peer] connection %s (remote: %s)")

// Peer represents a peer broker.
type Peer struct {
	sync.Mutex
	writer     *snappy.Writer // The writer to use for writing messages.
	reader     *snappy.Reader // The reader to use for reading messages.
	socket     net.Conn       // The underlying transport socket for reader and writers.
	frame      MessageFrame   // The current message frame.
	handshaken bool           // The flag for handshake.

	OnClosing   func()                            // Handler which is invoked when the peer is closing is received.
	OnHandshake func(*Peer, HandshakeEvent) error // Handler which is invoked when a handshake is received.
	OnMessage   func(Message)                     // Handler which is invoked when a new message is received.
}

// NewPeer creates a new peer for the connection.
func newPeer(conn net.Conn) *Peer {
	logging.Log(logConnection, "opened", conn.RemoteAddr().String())
	peer := &Peer{
		socket: conn,
		writer: snappy.NewBufferedWriter(conn),
		reader: snappy.NewReader(bufio.NewReaderSize(conn, readBufferSize)),
		frame:  make(MessageFrame, 0, 64),
	}

	// Spawn the send queue processor as well
	go peer.processSendQueue()
	return peer
}

// Send forwards the message to the remote server.
func (c *Peer) Send(channel []byte, payload []byte) error {
	c.Lock()
	defer c.Unlock()

	// Send simply appends the message to a frame
	c.frame = append(c.frame, Message{Channel: channel, Payload: payload})
	return nil
}

// processSendQueue flushes the current frame to the remote server
func (c *Peer) processSendQueue() {
	encoder := encoding.NewEncoder(c.writer)

	for {

		if len(c.frame) > 0 {
			// Encode the current frame
			c.Lock()
			err := encoder.Encode(c.frame)
			c.frame = c.frame[0:0]
			c.Unlock()

			// Something went wrong during the encoding
			if err != nil {
				logging.LogError("peer", "encoding frame", err)
			}

			// Flush the writer
			c.writer.Flush()
		}

		// Wait for a few milliseconds for the buffer to fill up and also let other goroutines
		// to be scheduled and run meanwhile.
		time.Sleep(5 * time.Millisecond)
	}
}

// Handshake sends a handshake message to the peer.
func (c *Peer) Handshake(node string) error {
	if c.handshaken {
		return nil // Avoid sending the handshake recursively.
	}

	c.handshaken = true
	if err := encoding.EncodeTo(c.writer, &HandshakeEvent{
		Key:  "",
		Node: node,
	}); err != nil {
		return err
	}

	// Flush the buffered writer so we'd actually write through the socket
	return c.writer.Flush()
}

// Process processes the messages.
func (c *Peer) Process() error {
	defer c.Close()
	decoder := encoding.NewDecoder(c.reader)

	// First message we need to decode is the handshake
	handshake, err := decodeHandshakeEvent(decoder)
	if err != nil {
		logging.LogError("peer", "decode handshake", err)
		return err
	}

	// Validate the handshake
	if err := c.OnHandshake(c, *handshake); err != nil {
		logging.LogError("peer", "handshake", err)
		return err
	}

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

// Close terminates the connection.
func (c *Peer) Close() error {
	logging.Log(logConnection, "closed", c.socket.RemoteAddr().String())

	// Close the peer.
	if c.OnClosing != nil {
		c.OnClosing()
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

/*

// NewPeerManager creates a new manager for the peers.
func NewPeerManager(service *Service) *PeerManager {
	return &PeerManager{
		service: service,
		peers:   collection.NewConcurrentMap(),
	}
}

// GetMember retrieves the member by its id
func (m *PeerManager) getMember(node string) *serf.Member {
	cluster := m.service.cluster
	for _, m := range cluster.Members() {
		if m.Name == node {
			return &m
		}
	}

	return nil
}

// Get retrieves a peer from the manager, if not find it attempts to connect to a peer.
func (m *PeerManager) Get(node string) (peer *Peer, ok bool) {
	// First we need to find the member
	member := m.getMember(node)
	if member == nil {
		return nil, false
	}

	// Attempt to retrieve the associated peer
	key := utils.GetHash([]byte(member.Name))
	peer, ok = m.peers.GetOrCreate(key, func() interface{} {
		addr := member.Tags["route"]
		if conn, err := net.Dial("tcp", addr); err != nil {
			return m.newPeer(conn)
		}
		return nil
	}).(*Peer)
	return
}
*/
