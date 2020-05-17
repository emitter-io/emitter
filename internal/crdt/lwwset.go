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

package crdt

import (
	"bytes"
	"sort"
	"sync"
	"time"

	"github.com/golang/snappy"
	"github.com/kelindar/binary"
	"github.com/kelindar/binary/nocopy"
)

// The expiration cutoff time for GCs
var gcCutoff = (6 * time.Hour).Nanoseconds()

// LWWTime represents a time pair.
type LWWTime struct {
	AddTime int64
	DelTime int64
}

// IsZero checks if the time is zero
func (t LWWTime) IsZero() bool {
	return (t.AddTime == 0 && t.DelTime == 0)
}

// IsAdded checks if add time is larger than remove time.
func (t LWWTime) IsAdded() bool {
	return t.AddTime != 0 && t.AddTime >= t.DelTime
}

// IsRemoved checks if remove time is larger than add time.
func (t LWWTime) IsRemoved() bool {
	return t.AddTime < t.DelTime
}

// IsExpired checks if the element was removed long time ago and can be safely garbage collected.
func (t LWWTime) isExpired() bool {
	return t.IsRemoved() && (t.DelTime+gcCutoff < Now())
}

// LWWSet represents a last-write-wins CRDT set.
type LWWSet struct {
	sync.Mutex
	Set LWWState
}

// NewLWWSet creates a new last-write-wins set with bias for 'add'.
func NewLWWSet() *LWWSet {
	return &LWWSet{
		Set: make(LWWState),
	}
}

// Add adds a value to the set.
func (s *LWWSet) Add(value string) {
	s.Lock()
	defer s.Unlock()

	v, _ := s.Set[value]
	s.Set[value] = LWWTime{AddTime: Now(), DelTime: v.DelTime}
}

// Remove removes the value from the set.
func (s *LWWSet) Remove(value string) {
	s.Lock()
	defer s.Unlock()

	v, _ := s.Set[value]
	s.Set[value] = LWWTime{AddTime: v.AddTime, DelTime: Now()}
}

// Contains checks if a value is present in the set.
func (s *LWWSet) Contains(value string) bool {
	s.Lock()
	defer s.Unlock()

	v, _ := s.Set[value]
	return v.IsAdded()
}

// Merge merges two LWW sets. This also modifies the set being merged in
// to leave only the delta.
func (s *LWWSet) Merge(r *LWWSet) {
	s.Lock()
	r.Lock()
	defer s.Unlock()
	defer r.Unlock()

	for key, rt := range r.Set {
		t, _ := s.Set[key]

		if t.AddTime < rt.AddTime {
			t.AddTime = rt.AddTime
		} else {
			rt.AddTime = 0 // Remove from delta
		}

		if t.DelTime < rt.DelTime {
			t.DelTime = rt.DelTime
		} else {
			rt.DelTime = 0 // Remove from delta
		}

		if rt.IsZero() {
			delete(r.Set, key) // Remove from delta
		} else {
			s.Set[key] = t  // Merge the new value
			r.Set[key] = rt // Update the delta
		}
	}
}

// Range iterates through the events for a specific prefix.
func (s *LWWSet) Range(prefix []byte, f func(string) bool) {
	s.Lock()
	defer s.Unlock()

	for k, v := range s.Set {
		if bytes.HasPrefix([]byte(k), prefix) && v.IsAdded() {
			if !f(k) { // If returns false, stop
				return
			}
		}
	}
}

// Clone copies a set into another set
func (s *LWWSet) Clone() *LWWSet {
	s.Lock()
	defer s.Unlock()

	items := make(LWWState, len(s.Set))
	for key, val := range s.Set {
		items[key] = val
	}
	return &LWWSet{Set: items}
}

// GC collects all the garbage in the set by simply removing it. This currently uses a very
// simplistic strategy with a static cutoff and expects all of the nodes in the cluster to
// have a relatively synchronized time, in absense of which this would cause inconsistency.
func (s *LWWSet) GC() {
	s.Lock()
	defer s.Unlock()

	for key, val := range s.Set {
		if val.isExpired() {
			delete(s.Set, key)
		}
	}
}

// The clock for unit-testing
type clock func() int64

// Now gets the current time in Unix nanoseconds
var Now clock = func() int64 {
	return time.Now().UnixNano()
}

// ------------------------------------------------------------------------------------

// LWWState represents the internal state
type LWWState = map[string]LWWTime

type entry struct {
	Value   nocopy.String
	AddTime int64
	DelTime int64
}

// Marshal marshals the state into a compresed buffer.
func (s *LWWSet) Marshal() []byte {
	s.Lock()
	defer s.Unlock()

	count, breakout := 0, 100000
	msg := make([]entry, 0, breakout)
	for k, v := range s.Set {
		msg = append(msg, entry{
			Value:   nocopy.String(k),
			AddTime: v.AddTime,
			DelTime: v.DelTime,
		})

		// Since we're iterating over a map, the iteration should be done in pseudo-random
		// order. Hence, we take advantage of this and break at 100K subscriptions in order
		// to make sure the gossip message fits under 10MB (max size).
		if count++; count >= breakout {
			break
		}
	}

	// Sort the values in order to compress more
	sort.Slice(msg, func(i, j int) bool {
		return msg[i].Value < msg[j].Value
	})

	buf, err := binary.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return snappy.Encode(nil, buf)
}

// Unmarshal unmarshals the encoded state.
func (s *LWWSet) Unmarshal(b []byte) error {
	s.Lock()
	defer s.Unlock()

	if s.Set == nil {
		s.Set = make(LWWState, 0)
	}

	decoded, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}

	var msg []entry
	if err := binary.Unmarshal(decoded, &msg); err != nil {
		return err
	}

	for _, v := range msg {
		s.Set[string(v.Value)] = LWWTime{
			AddTime: v.AddTime,
			DelTime: v.DelTime,
		}
	}
	return nil
}
