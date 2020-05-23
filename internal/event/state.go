/**********************************************************************************
* Copyright (c) 2009-2020 Misakai Ltd.
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

package event

import (
	bin "encoding/binary"

	"github.com/emitter-io/emitter/internal/event/crdt"
	"github.com/kelindar/binary"
	"github.com/weaveworks/mesh"
)

// Time represents an event time (CRDT metadata).
type Time = crdt.Time

// State represents globally synchronised state.
type State struct {
	m map[uint8]crdt.Set
}

// NewState creates a new replicated state.
func NewState() *State {
	return &State{
		m: map[uint8]crdt.Set{
			typeSubscription: *crdt.New(),
		},
	}
}

// DecodeState decodes the replicated state.
func DecodeState(buf []byte) (*State, error) {
	out := NewState()
	err := binary.Unmarshal(buf, &out.m)
	return out, err
}

// Encode serializes our complete state to a slice of byte-slices.
func (st *State) Encode() [][]byte {
	for _, set := range st.m {
		set.GC()
	}

	// Encode the byte map
	encoded, _ := binary.Marshal(st.m)
	return [][]byte{encoded}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete state.
func (st *State) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	otherState := other.(*State)
	for typ, lww := range otherState.m {
		otherLww := st.m[typ] // Get the corresponding set to merge with
		lww.Merge(&otherLww)  // Merges and changes otherState to be a delta
	}

	// Return the delta after merging
	return otherState
}

// Add adds the unit to the state.
func (st *State) Add(ev Event) {
	set := st.m[ev.unitType()]
	set.Add(ev.Encode())
}

// Remove removes the unit from the state.
func (st *State) Remove(ev Event) {
	set := st.m[ev.unitType()]
	set.Remove(ev.Encode())
}

// Subscriptions iterates through all of the subscription units. This call is
// blocking and will lock the entire set of subscriptions while iterating.
func (st *State) Subscriptions(f func(Subscription, Time)) {
	set := st.m[typeSubscription]
	set.Range(nil, func(v string, t crdt.Time) bool {
		if ev, err := decodeSubscription(v); err == nil {
			f(ev, t)
		}
		return true
	})
}

// SubscriptionsOf iterates through the subscription events for a specific peer.
func (st *State) SubscriptionsOf(name mesh.PeerName, f func(Subscription)) {
	buffer := make([]byte, 10, 10)
	offset := bin.PutUvarint(buffer, uint64(name))
	prefix := buffer[:offset]

	// Copy since the Range() is locked
	var events []Subscription
	set := st.m[typeSubscription]
	set.Range(prefix, func(v string, t crdt.Time) bool {
		if t.IsAdded() {
			if ev, err := decodeSubscription(v); err == nil {
				events = append(events, ev)
			}
		}
		return true
	})

	// Invoke the callback, without blocking the state
	for _, v := range events {
		f(v)
	}
}
