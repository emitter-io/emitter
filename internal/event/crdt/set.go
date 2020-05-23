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
	"sync"
	"time"
)

// The expiration cutoff time for GCs
var gcCutoff = (6 * time.Hour).Nanoseconds()

// Time represents a time pair.
type Time struct {
	AddTime int64
	DelTime int64
}

// IsZero checks if the time is zero
func (t Time) IsZero() bool {
	return (t.AddTime == 0 && t.DelTime == 0)
}

// IsAdded checks if add time is larger than remove time.
func (t Time) IsAdded() bool {
	return t.AddTime != 0 && t.AddTime >= t.DelTime
}

// IsRemoved checks if remove time is larger than add time.
func (t Time) IsRemoved() bool {
	return t.AddTime < t.DelTime
}

// IsExpired checks if the element was removed long time ago and can be safely garbage collected.
func (t Time) isExpired() bool {
	return t.IsRemoved() && (t.DelTime+gcCutoff < Now())
}

// ------------------------------------------------------------------------------------

// Set represents a last-write-wins CRDT set.
type Set struct {
	lock *sync.Mutex     // The associated mutex
	data map[string]Time // The data containing the set
}

// New creates a new last-write-wins set with bias for 'add'.
func New() *Set {
	return NewWith(make(map[string]Time, 64))
}

// NewWith creates a new last-write-wins set with bias for 'add'.
func NewWith(items map[string]Time) *Set {
	return &Set{
		lock: new(sync.Mutex),
		data: items,
	}
}

// Add adds a value to the set.
func (s *Set) Add(value string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, _ := s.data[value]
	s.data[value] = Time{AddTime: Now(), DelTime: v.DelTime}
}

// Remove removes the value from the set.
func (s *Set) Remove(value string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, _ := s.data[value]
	s.data[value] = Time{AddTime: v.AddTime, DelTime: Now()}
}

// Contains checks if a value is present in the set.
func (s *Set) Contains(value string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, _ := s.data[value]
	return v.IsAdded()
}

// Merge merges two LWW sets. This also modifies the set being merged in
// to leave only the delta.
func (s *Set) Merge(r *Set) {
	s.lock.Lock()
	r.lock.Lock()
	defer s.lock.Unlock()
	defer r.lock.Unlock()

	for key, rt := range r.data {
		t, _ := s.data[key]

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
			delete(r.data, key) // Remove from delta
		} else {
			s.data[key] = t  // Merge the new value
			r.data[key] = rt // Update the delta
		}
	}
}

// Range iterates through the events for a specific prefix.
func (s *Set) Range(prefix []byte, f func(string, Time) bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for k, v := range s.data {
		if bytes.HasPrefix([]byte(k), prefix) {
			if !f(k, v) { // If returns false, stop
				return
			}
		}
	}
}

// Clone copies a set into another set
func (s *Set) Clone() *Set {
	s.lock.Lock()
	defer s.lock.Unlock()

	items := make(map[string]Time, len(s.data))
	for key, val := range s.data {
		items[key] = val
	}
	return &Set{
		lock: new(sync.Mutex),
		data: items,
	}
}

// GC collects all the garbage in the set by simply removing it. This currently uses a very
// simplistic strategy with a static cutoff and expects all of the nodes in the cluster to
// have a relatively synchronized time, in absense of which this would cause inconsistency.
func (s *Set) GC() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for key, val := range s.data {
		if val.isExpired() {
			delete(s.data, key)
		}
	}
}

// The clock for unit-testing
type clock func() int64

// Now gets the current time in Unix nanoseconds
var Now clock = func() int64 {
	return time.Now().UnixNano()
}
