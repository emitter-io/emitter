/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
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
	"sync"
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/internal/async"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/weaveworks/mesh"
)

// Peer implements subscription.Subscriber
var _ message.Subscriber = &Peer{}

const (
	defaultFrameSize = 128              // Default message frame size to use
	maxByteFrameSize = 10 * 1024 * 1024 // Hard limit imposed by our underlying gossip
)

// Peer represents a remote peer.
type Peer struct {
	sync.Mutex
	sender   mesh.Gossip        // The gossip interface to use for sending.
	name     mesh.PeerName      // The peer name for communicating.
	frame    message.Frame      // The current message frame.
	subs     *message.Counters  // The SSIDs of active subscriptions for this peer.
	activity int64              // The time of last activity of the peer.
	cancel   context.CancelFunc // The cancellation function.
}

// NewPeer creates a new peer for the connection.
func (s *Swarm) newPeer(name mesh.PeerName) *Peer {
	peer := &Peer{
		sender:   s.gossip,
		name:     name,
		frame:    message.NewFrame(defaultFrameSize),
		subs:     message.NewCounters(),
		activity: time.Now().Unix(),
	}

	// Spawn the send queue processor
	peer.cancel = async.Repeat(context.Background(), 5*time.Millisecond, peer.processSendQueue)
	return peer
}

// Occurs when the peer is subscribed
func (p *Peer) onSubscribe(encodedEvent string, ssid message.Ssid) bool {
	return p.subs.Increment(ssid, []byte(encodedEvent))
}

// Occurs when the peer is unsubscribed
func (p *Peer) onUnsubscribe(encodedEvent string, ssid message.Ssid) bool {
	return p.subs.Decrement(ssid)
}

// Close termintes the peer and stops everything associated with this peer.
func (p *Peer) Close() error {
	p.Lock()
	defer p.Unlock()

	if p.cancel != nil {
		p.cancel()
	}

	return nil
}

// ID returns the unique identifier of the subsriber.
func (p *Peer) ID() string {
	return p.name.String()
}

// Type returns the type of the subscriber.
func (p *Peer) Type() message.SubscriberType {
	return message.SubscriberRemote
}

// IsActive checks whether a peer is still active or not.
func (p *Peer) IsActive() bool {
	return (atomic.LoadInt64(&p.activity) + 30) > time.Now().Unix()
}

// Send forwards the message to the remote server.
func (p *Peer) Send(m *message.Message) error {
	p.Lock()
	defer p.Unlock()

	// Make sure we don't send to a dead peer
	if p.IsActive() {
		p.frame = append(p.frame, *m)
	}

	return nil
}

// swap swaps the frame and returns the frame we can encode.
func (p *Peer) swap() (swapped message.Frame) {
	p.Lock()
	defer p.Unlock()

	swapped = p.frame
	p.frame = message.NewFrame(defaultFrameSize)
	return
}

// processSendQueue flushes the current frame to the remote server
func (p *Peer) processSendQueue() {
	if len(p.frame) == 0 {
		return
	}

	// Swap the frame and split the frame in chunks of at most 10MB
	// for gossip unicast to work.
	frame := p.swap()
	for {
		var chunk message.Frame
		chunk, frame = frame.Split(maxByteFrameSize)
		if len(chunk) == 0 {
			break
		}

		buffer := chunk.Encode()
		if err := p.sender.GossipUnicast(p.name, buffer); err != nil {
			logging.LogError("peer", "gossip unicast", err)
		}
	}
}

// ------------------------------------------------------------------------------------

// DeadPeer represents a peer which is no longer online
type deadPeer struct {
	name mesh.PeerName
}

// ID returns the unique identifier of the subsriber.
func (p *deadPeer) ID() string {
	return p.name.String()
}

// Type returns the type of the subscriber.
func (p *deadPeer) Type() message.SubscriberType {
	return message.SubscriberOffline
}

// Send forwards the message to the remote server.
func (p *deadPeer) Send(m *message.Message) error {
	return nil
}
