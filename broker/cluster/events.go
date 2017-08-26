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
	"github.com/emitter-io/emitter/collection"
	"github.com/emitter-io/emitter/encoding"
	"github.com/emitter-io/emitter/security"
	"github.com/weaveworks/mesh"
)

// Subscription represents a subscription.
type Subscription struct {
	Ssid    []uint32 // The Ssid of the message
	Channel []byte   // The channel of the message
}

// MessageFrame represents a message frame which is sent through the wire to the
// remote server and contains a set of messages
type MessageFrame []*Message

// Message represents a message which has to be routed.
type Message struct {
	Ssid    []uint32 // The Ssid of the message
	Channel []byte   // The channel of the message
	Payload []byte   // The payload of the message
}

// decodeMessageFrame decodes the message frame from the decoder.
func decodeMessageFrame(decoder encoding.Decoder) (out MessageFrame, err error) {
	out = make(MessageFrame, 0, 64)
	err = decoder.Decode(&out)
	return
}

// SubscriptionEvent represents a subscription event.
type SubscriptionEvent struct {
	Peer mesh.PeerName // The name of the peer.
	Conn security.ID   // The connection identifier.
	Ssid []uint32      // The SSID for the subscription.
}

// Encode encodes the event to byte representation.
func (e *SubscriptionEvent) Encode() []byte {
	buf, err := encoding.Encode(e)
	if err != nil {
		panic(err)
	}

	return buf
}

// decodeSubscriptionEvent decodes the event
func decodeSubscriptionEvent(buf []byte) (out *SubscriptionEvent, err error) {
	out = &SubscriptionEvent{}
	err = encoding.Decode(buf, out)
	return
}

// SubscriptionState represents globally synchronised state.
type subscriptionState collection.LWWSet

// newSubscriptionState creates a new last-write-wins set with bias for 'add'.
func newSubscriptionState() *subscriptionState {
	return (*subscriptionState)(collection.NewLWWSet())
}

// decodeSubscriptionState decodes the state
func decodeSubscriptionState(buf []byte) (*subscriptionState, error) {
	out := map[interface{}]collection.LWWTime{}
	err := encoding.Decode(buf, &out)
	return &subscriptionState{Set: out}, err
}

// Encode serializes our complete state to a slice of byte-slices.
func (st *subscriptionState) Encode() [][]byte {
	lww := (*collection.LWWSet)(st)
	lww.Lock()
	defer lww.Unlock()

	buf, err := encoding.Encode(lww.Set)
	if err != nil {
		panic(err)
	}

	return [][]byte{buf}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete state.
func (st *subscriptionState) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	lww := (*collection.LWWSet)(st)

	otherState := other.(*subscriptionState)
	otherLww := (*collection.LWWSet)(otherState)

	lww.Merge(otherLww) // Merges and changes otherState to be a delta
	return otherState   // Return the delta after merging
}

// Add adds the subscription event to the state.
func (st *subscriptionState) Add(ev string) {
	(*collection.LWWSet)(st).Add(ev)
}

// Remove removes the subscription event from the state.
func (st *subscriptionState) Remove(ev string) {
	(*collection.LWWSet)(st).Remove(ev)
}

// All ...
func (st *subscriptionState) All() []interface{} {
	return (*collection.LWWSet)(st).All()
}
