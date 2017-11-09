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
	"sync"
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/utils"
	"github.com/weaveworks/mesh"
)

// Peer implements subscription.Subscriber
var _ subscription.Subscriber = &Peer{}

// Peer represents a remote peer.
type Peer struct {
	sync.Mutex
	sender   mesh.Gossip            // The gossip interface to use for sending.
	name     mesh.PeerName          // The peer name for communicating.
	frame    message.Frame          // The current message frame.
	subs     *subscription.Counters // The SSIDs of active subscriptions for this peer.
	activity int64                  // The time of last activity of the peer.
	closing  chan bool              // The closing channel for the peer.
}

// NewPeer creates a new peer for the connection.
func (s *Swarm) newPeer(name mesh.PeerName) *Peer {
	peer := &Peer{
		sender:   s.gossip,
		name:     name,
		frame:    make(message.Frame, 0, 64),
		subs:     subscription.NewCounters(),
		activity: time.Now().Unix(),
		closing:  make(chan bool),
	}

	// Spawn the send queue processor
	utils.Repeat(peer.processSendQueue, 5*time.Millisecond, peer.closing)
	return peer
}

// Occurs when the peer is subscribed
func (p *Peer) onSubscribe(encodedEvent string, ssid subscription.Ssid) bool {
	return p.subs.Increment(ssid, []byte(encodedEvent))
}

// Occurs when the peer is unsubscribed
func (p *Peer) onUnsubscribe(encodedEvent string, ssid subscription.Ssid) bool {
	return p.subs.Decrement(ssid)
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
		p.frame.Append(0, ssid, channel, payload)
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

	// Encode the current frame
	p.Lock()
	buffer, err := p.frame.Encode()
	p.frame = p.frame[:0]
	p.Unlock()

	// Log the error
	if err != nil {
		logging.LogError("peer", "encoding frame", err)
		return
	}

	// Send the frame directly to the peer.
	if err := p.sender.GossipUnicast(p.name, buffer); err != nil {
		logging.LogError("peer", "gossip unicast", err)
	}

}
