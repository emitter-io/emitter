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
	bin "encoding/binary"

	"github.com/emitter-io/emitter/internal/crdt"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/kelindar/binary"
	"github.com/kelindar/binary/nocopy"
	"github.com/weaveworks/mesh"
)

// SubscriptionEvent represents a subscription event.
type SubscriptionEvent struct {
	Peer    mesh.PeerName // The name of the peer. This must be first, since we're doing prefix search.
	Conn    security.ID   // The connection identifier.
	User    nocopy.String // The connection username.
	Channel nocopy.Bytes  // The channel string.
	Ssid    message.Ssid  // The SSID for the subscription.
}

// ConnID returns globally-unique identifier for the connection.
func (e *SubscriptionEvent) ConnID() string {
	return e.Conn.Unique(uint64(e.Peer), "emitter")
}

// Encode encodes the event to string representation.
func (e *SubscriptionEvent) Encode() string {
	buf, _ := binary.Marshal(e)
	return string(buf)
}

// decodeSubscriptionEvent decodes the event
func decodeSubscriptionEvent(encoded string) (SubscriptionEvent, error) {
	var out SubscriptionEvent
	return out, binary.Unmarshal([]byte(encoded), &out)
}

// ------------------------------------------------------------------------------------

// SubscriptionState represents globally synchronised state.
type subscriptionState crdt.LWWSet

// newSubscriptionState creates a new last-write-wins set with bias for 'add'.
func newSubscriptionState() *subscriptionState {
	return (*subscriptionState)(crdt.NewLWWSet())
}

// decodeSubscriptionState decodes the state
func decodeSubscriptionState(buf []byte) (*subscriptionState, error) {
	out := crdt.NewLWWSet()
	if err := out.Unmarshal(buf); err != nil {
		return nil, err
	}

	return (*subscriptionState)(out), nil
}

// Encode serializes our complete state to a slice of byte-slices.
func (st *subscriptionState) Encode() [][]byte {
	lww := (*crdt.LWWSet)(st)
	lww.GC()

	return [][]byte{lww.Marshal()}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete state.
func (st *subscriptionState) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	lww := (*crdt.LWWSet)(st)

	otherState := other.(*subscriptionState)
	otherLww := (*crdt.LWWSet)(otherState)

	lww.Merge(otherLww) // Merges and changes otherState to be a delta
	return otherState   // Return the delta after merging
}

// Add adds the subscription event to the state.
func (st *subscriptionState) Add(ev string) {
	(*crdt.LWWSet)(st).Add(ev)
}

// Remove removes the subscription event from the state.
func (st *subscriptionState) Remove(ev string) {
	(*crdt.LWWSet)(st).Remove(ev)
}

// Range iterates through the events for a specific peer.
func (st *subscriptionState) Range(name mesh.PeerName, f func(SubscriptionEvent)) {
	buffer := make([]byte, 10, 10)
	offset := bin.PutUvarint(buffer, uint64(name))
	prefix := buffer[:offset]

	// Copy since the Range() is locked
	var events []SubscriptionEvent
	(*crdt.LWWSet)(st).Range(prefix, func(v string) bool {
		if ev, err := decodeSubscriptionEvent(v); err == nil {
			events = append(events, ev)
		}
		return true
	})

	// Invoke the callback, without blocking the state
	for _, v := range events {
		f(v)
	}
}

// Clone clones the state
func (st *subscriptionState) Clone() crdt.LWWState {
	return (*crdt.LWWSet)(st).Clone().Set
}
