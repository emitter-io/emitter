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

package collection

import (
	"sync"
	"time"
)

// LWWState represents the internal state
type LWWState = map[string]LWWTime

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
	defer s.Unlock()

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

// All gets all items in the set.
func (s *LWWSet) All() LWWState {
	s.Lock()
	defer s.Unlock()

	items := make(LWWState, len(s.Set))
	for key, val := range s.Set {
		items[key] = val
	}
	return items
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
