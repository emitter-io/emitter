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
	"io"
	"path"

	"github.com/emitter-io/emitter/internal/event/crdt"
	"github.com/golang/snappy"
	"github.com/kelindar/binary"
	"github.com/weaveworks/mesh"
)

// Value represents an event time & value.
type Value = crdt.Value

// State represents globally synchronised state.
type State struct {
	durable bool               // Whether the state is durable or not.
	subsets map[uint8]crdt.Map // The subsets of the state.
}

// NewState creates a new replicated state.
func NewState(dir string) *State {
	durable := dir != ""
	return &State{
		durable: durable,
		subsets: map[uint8]crdt.Map{
			typeSub:  crdt.New(durable, ""),
			typeBan:  crdt.New(durable, fileOf(dir, "ban.db")),
			typeConn: crdt.New(durable, ""),
		},
	}
}

// FileOf creates a filename for the specific directory
func fileOf(dir, name string) string {
	if dir == ":memory:" {
		return dir
	}

	return path.Join(dir, "ban.db")
}

// DecodeState decodes the replicated state.
func DecodeState(buf []byte) (out *State, err error) {

	// Decode the state, while decoding it can only be volatile (as per use-case)
	decoded := make(map[uint8]crdt.Volatile)
	if buf, err = snappy.Decode(nil, buf); err == nil {
		err = binary.Unmarshal(buf, &decoded)
	}

	// Copy the volatile set into the state
	out = NewState("")
	for typ, set := range decoded {
		set := set // Make sure to not add the iterator
		out.subsets[typ] = &set
	}

	return
}

// Encode serializes our complete state to a slice of byte-slices.
func (st *State) Encode() [][]byte {
	if st.durable {
		subsets := make(map[uint8]crdt.Durable)
		for k, v := range st.subsets {
			subsets[k] = *v.(*crdt.Durable)
		}

		encoded, _ := binary.Marshal(subsets)
		return [][]byte{snappy.Encode(nil, encoded)}
	}

	subsets := make(map[uint8]crdt.Volatile)
	for k, v := range st.subsets {
		subsets[k] = *v.(*crdt.Volatile)
	}

	encoded, _ := binary.Marshal(subsets)
	return [][]byte{snappy.Encode(nil, encoded)}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete state.
func (st *State) Merge(other mesh.GossipData) mesh.GossipData {
	count := 0
	otherState := other.(*State)
	for typ, lww := range st.subsets {
		otherLww := otherState.subsets[typ] // Get the corresponding set to merge with
		lww.Merge(otherLww)                 // Merges and changes otherState to be a delta
		count += otherLww.Count()
	}

	// If nothing is new, return nil state (otherwise gossip will go nuts with 3+ brokers)
	if count == 0 {
		return nil
	}

	// Return the delta after merging
	return otherState
}

// Add adds the unit to the state.
func (st *State) Add(ev Event) {
	set := st.subsets[ev.unitType()]
	set.Add(ev.Key(), ev.Val())
}

// Del removes the unit from the state.
func (st *State) Del(ev Event) {
	set := st.subsets[ev.unitType()]
	set.Del(ev.Key())
}

// Has checks if the state contains an event.
func (st *State) Has(ev Event) bool {
	set := st.subsets[ev.unitType()]
	return set.Has(ev.Key())
}

// Subscriptions iterates through all of the subscription units. This call is
// blocking and will lock the entire set of subscriptions while iterating.
func (st *State) Subscriptions(f func(*Subscription, Value)) {
	set := st.subsets[typeSub]
	set.Range(nil, true, func(v string, t Value) bool {
		if ev, err := decodeSubscription(v, t.Value()); err == nil {
			f(&ev, t)
		}
		return true
	})
}

// SubscriptionsOf iterates through the subscription events for a specific peer.
func (st *State) SubscriptionsOf(name mesh.PeerName, f func(*Subscription)) {
	for k, v := range st.findEventsOf(typeSub, prefixOf(name), false) {
		if ev, err := decodeSubscription(k, v.Value()); err == nil {
			f(&ev)
		}
	}
}

// ConnectionsOf iterates through the connection events for a specific peer.
func (st *State) ConnectionsOf(name mesh.PeerName, f func(*Connection)) {
	for k, v := range st.findEventsOf(typeConn, prefixOf(name), false) {
		if ev, err := decodeConnection(k, v.Value()); err == nil {
			f(&ev)
		}
	}
}

// findEventsOf ranges over the events of a specific type and copies them for concurrent usage.
func (st *State) findEventsOf(typ uint8, prefix []byte, tombstones bool) map[string]Value {
	events := make(map[string]Value)
	st.subsets[typ].Range(prefix, tombstones, func(k string, v crdt.Value) bool {
		events[k] = v
		return true
	})
	return events
}

// Close closes the set gracefully.
func (st *State) Close() error {
	for _, set := range st.subsets {
		if closer, ok := set.(io.Closer); ok {
			closer.Close()
		}
	}
	return nil
}

// prefixOf returns a binary prefix for search.
func prefixOf(name mesh.PeerName) []byte {
	prefix := make([]byte, 8)
	binary.BigEndian.PutUint64(prefix, uint64(name))
	return prefix
}
