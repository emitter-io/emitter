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
	"bytes"
	bin "encoding/binary"
	"github.com/golang/snappy"

	"github.com/emitter-io/emitter/internal/collection"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/kelindar/binary"
	"github.com/weaveworks/mesh"
)

// SubscriptionEvent represents a subscription event.
type SubscriptionEvent struct {
	Peer mesh.PeerName // The name of the peer.
	Conn security.ID   // The connection identifier.
	Ssid message.Ssid  // The SSID for the subscription.
}

// Encode encodes the event to string representation.
func (e *SubscriptionEvent) Encode() string {

	// Prepare a buffer and leave some space, since we're encoding in varint
	buf := make([]byte, 20+(6*len(e.Ssid)))
	offset := 0

	// Encode everything as variable-size unsigned integers to save space
	offset += bin.PutUvarint(buf[offset:], uint64(e.Peer))
	offset += bin.PutUvarint(buf[offset:], uint64(e.Conn))
	for _, ssidPart := range e.Ssid {
		offset += bin.PutUvarint(buf[offset:], uint64(ssidPart))
	}

	return string(buf[:offset])
}

// decodeSubscriptionEvent decodes the event
func decodeSubscriptionEvent(encoded string) (SubscriptionEvent, error) {
	out := SubscriptionEvent{}
	buf := []byte(encoded)

	reader := bytes.NewReader(buf)

	// Read the peer name
	peer, err := bin.ReadUvarint(reader)
	out.Peer = mesh.PeerName(peer)
	if err != nil {
		return out, err
	}

	// Read the connection identifier
	conn, err := bin.ReadUvarint(reader)
	out.Conn = security.ID(conn)
	if err != nil {
		return out, err
	}

	// Read the SSID until we're finished
	out.Ssid = make([]uint32, 0, 2)
	for reader.Len() > 0 {
		ssidPart, err := bin.ReadUvarint(reader)
		out.Ssid = append(out.Ssid, uint32(ssidPart))
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

// ------------------------------------------------------------------------------------

// SubscriptionState represents globally synchronised state.
type subscriptionState collection.LWWSet

// newSubscriptionState creates a new last-write-wins set with bias for 'add'.
func newSubscriptionState() *subscriptionState {
	return (*subscriptionState)(collection.NewLWWSet())
}

// decodeSubscriptionState decodes the state
func decodeSubscriptionState(buf []byte) (*subscriptionState, error) {
	out, err := decodeState(buf)
	if err != nil {
		return nil, err
	}

	return &subscriptionState{Set: out}, err
}

// Encode serializes our complete state to a slice of byte-slices.
func (st *subscriptionState) Encode() [][]byte {
	lww := (*collection.LWWSet)(st)
	lww.GC()
	lww.Lock()
	defer lww.Unlock()

	encoded := encodeState(lww.Set)
	return [][]byte{encoded}
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

// RemoveAll removes all of the subscription events by prefix.
func (st *subscriptionState) RemoveAll(name mesh.PeerName) {
	buffer := make([]byte, 10, 10)
	offset := bin.PutUvarint(buffer, uint64(name))
	prefix := buffer[:offset]

	for ev, v := range st.All() {
		if bytes.HasPrefix([]byte(ev), prefix) && v.IsAdded() {
			st.Remove(ev)
		}
	}
}

// All ...
func (st *subscriptionState) All() collection.LWWState {
	return (*collection.LWWSet)(st).All()
}

// ------------------------------------------------------------------------------------

type state struct {
	Keys []string
	Adds []int64
	Dels []int64
}

func encodeState(s collection.LWWState) []byte {
	msg := state{
		Keys: make([]string, 0, len(s)),
		Adds: make([]int64, 0, len(s)),
		Dels: make([]int64, 0, len(s)),
	}

	var count int
	var addOffset, delOffset int64
	for k, v := range s {
		msg.Keys = append(msg.Keys, k)
		msg.Adds = append(msg.Adds, v.AddTime-addOffset)
		msg.Dels = append(msg.Dels, v.DelTime-delOffset)

		if addOffset == 0 {
			addOffset = v.AddTime
			delOffset = v.DelTime
		}

		// Since we're iterating over a map, the iteration should be done in pseudo-random
		// order. Hence, we take advantage of this and break at 100K subscriptions in order
		// to make sure the gossip message fits under 10MB (max size).
		if count++; count >= 100000 {
			break
		}
	}

	buf, err := binary.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return snappy.Encode(nil, buf)
}

func decodeState(b []byte) (collection.LWWState, error) {
	decoded, err := snappy.Decode(nil, b)
	if err != nil {
		return nil, err
	}

	var msg state
	if err := binary.Unmarshal(decoded, &msg); err != nil {
		return nil, err
	}

	var addOffset, delOffset int64
	out := make(collection.LWWState, len(msg.Keys))
	for i := 0; i < len(msg.Keys); i++ {
		k := msg.Keys[i]
		addTime := msg.Adds[i] + addOffset
		delTime := msg.Dels[i] + delOffset

		if addOffset == 0 {
			addOffset = addTime
			delOffset = delTime
		}

		out[k] = collection.LWWTime{
			AddTime: addTime,
			DelTime: delTime,
		}
	}
	return out, nil
}
