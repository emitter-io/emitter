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
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/emitter-io/emitter/encoding"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/utils"
	"github.com/golang/snappy"
	"github.com/weaveworks/mesh"
)

// Peer implements subscription.Subscriber
var _ subscription.Subscriber = &Peer{}

// Peer represents a remote peer.
type Peer struct {
	sync.Mutex
	sender   mesh.Gossip                  // The gossip interface to use for sending.
	name     mesh.PeerName                // The peer name for communicating.
	frame    MessageFrame                 // The current message frame.
	subs     map[string]subscription.Ssid // The SSIDs of active subscriptions for this peer.
	activity int64                        // The time of last activity of the peer.
	closing  chan bool                    // The closing channel for the peer.
}

// NewPeer creates a new peer for the connection.
func (s *Swarm) newPeer(name mesh.PeerName) *Peer {
	peer := &Peer{
		sender:   s.gossip,
		name:     name,
		frame:    make(MessageFrame, 0, 64),
		subs:     make(map[string]subscription.Ssid),
		activity: time.Now().Unix(),
		closing:  make(chan bool),
	}

	// Spawn the send queue processor
	utils.Repeat(peer.processSendQueue, 5*time.Millisecond, peer.closing)
	return peer
}

// Occurs when the peer is subscribed
func (p *Peer) onSubscribe(encodedEvent string, ssid subscription.Ssid) {
	p.Lock()
	defer p.Unlock()

	p.subs[encodedEvent] = ssid
}

// Occurs when the peer is unsubscribed
func (p *Peer) onUnsubscribe(encodedEvent string, ssid subscription.Ssid) {
	p.Lock()
	defer p.Unlock()

	delete(p.subs, encodedEvent)
}

// Close termintes the peer and stops everything associated with this peer.
func (p *Peer) Close() {
	p.Lock()
	defer p.Unlock()

	close(p.closing)
}

// ID returns the unique identifier of the subsriber.
func (p *Peer) ID() string {
	return p.name.String()
}

// Type returns the type of the subscriber.
func (p *Peer) Type() subscription.SubscriberType {
	return subscription.SubscriberRemote
}

// IsActive checks whether a peer is still active or not.
func (p *Peer) IsActive() bool {
	return (atomic.LoadInt64(&p.activity) + 30) > time.Now().Unix()
}

// Send forwards the message to the remote server.
func (p *Peer) Send(ssid subscription.Ssid, channel []byte, payload []byte) error {
	p.Lock()
	defer p.Unlock()

	// TODO: Make sure we don't send to a dead peer
	if p.IsActive() {

		// Send simply appends the message to a frame
		p.frame = append(p.frame, &Message{Ssid: ssid, Channel: channel, Payload: payload})

	}

	return nil
}

// Touch updates the activity time of the peer.
func (p *Peer) touch() {
	atomic.StoreInt64(&p.activity, time.Now().Unix())
}

// processSendQueue flushes the current frame to the remote server
func (p *Peer) processSendQueue() {
	if len(p.frame) == 0 {
		return // Nothing to send.
	}

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
	if err := snappy.Close(); err == nil {
		if err := p.sender.GossipUnicast(p.name, buffer.Bytes()); err != nil {
			//logging.LogError("peer", "gossip unicast", err)
		}
	}
}
