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
	"reflect"
	"sync"

	"github.com/kelindar/binary"
)

// Volatile represents a last-write-wins CRDT set.
type Volatile struct {
	lock *sync.Mutex      // The associated mutex
	data map[string]Value // The data containing the set
}

// NewVolatile creates a new last-write-wins set with bias for 'add'.
func NewVolatile() *Volatile {
	return newVolatileWith(make(map[string]Value, 64))
}

// newVolatileWith creates a new last-write-wins set with bias for 'add'.
func newVolatileWith(items map[string]Value) *Volatile {
	return &Volatile{
		lock: new(sync.Mutex),
		data: items,
	}
}

// Fetch fetches the item from the dictionary.
func (s *Volatile) fetch(item string) Value {
	if t, ok := s.data[item]; ok {
		return t
	}
	return newValue()
}

// Add adds a value to the set.
func (s *Volatile) Add(item string, value []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()

	t, now := s.fetch(item), Now()
	if t.AddTime() < now {
		t.setAddTime(now)
		t.setValue(value)
		s.data[item] = t
	}
}

// Del removes the value from the set.
func (s *Volatile) Del(item string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	t, now := s.fetch(item), Now()
	if t.DelTime() < now {
		t.setDelTime(now)
		s.data[item] = t
	}
}

// Has checks if a value is present in the set.
func (s *Volatile) Has(item string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	t := s.fetch(item)
	return t.IsAdded()
}

// Get retrieves the time for an item.
func (s *Volatile) Get(item string) Value {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.fetch(item)
}

// Merge merges two LWW sets. This also modifies the set being merged in
// to leave only the delta.
func (s *Volatile) Merge(other Map) {
	r := other.(*Volatile)
	s.lock.Lock()
	r.lock.Lock()
	defer s.lock.Unlock()
	defer r.lock.Unlock()

	for key, rt := range r.data {
		st := s.fetch(key)

		// Update add time & value
		if st.AddTime() < rt.AddTime() {
			st.setAddTime(rt.AddTime())
		} else {
			rt.setAddTime(0) // Remove from delta
		}

		// Update delete time
		if st.DelTime() < rt.DelTime() {
			st.setDelTime(rt.DelTime())
		} else {
			rt.setDelTime(0) // Remove from delta
		}

		if rt.IsZero() {
			delete(r.data, key) // Remove from delta
		} else {
			st.setValue(rt.Value()) // Set the new value
			s.data[key] = st        // Merge the new value
			r.data[key] = rt        // Update the delta
		}
	}
}

// Range iterates through the events for a specific prefix.
func (s *Volatile) Range(prefix []byte, tombstones bool, f func(string, Value) bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for k, v := range s.data {
		if !bytes.HasPrefix(binary.ToBytes(k), prefix) {
			continue
		}

		if tombstones || v.IsAdded() {
			if !f(k, v) { // If returns false, stop
				return
			}
		}
	}
}

// Count returns the number of items in the set.
func (s *Volatile) Count() (count int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return len(s.data)
}

// GetBinaryCodec retrieves a custom binary codec.
func (s *Volatile) GetBinaryCodec() binary.Codec {
	return new(codecVolatile)
}

// ------------------------------------------------------------------------------------

type codecVolatile struct{}

// Encode encodes a value into the encoder.
func (c *codecVolatile) EncodeTo(e *binary.Encoder, rv reflect.Value) (err error) {
	s := rv.Interface().(Volatile)
	s.lock.Lock()
	defer s.lock.Unlock()

	e.WriteUvarint(uint64(len(s.data)))
	for k, t := range s.data {
		e.WriteString(k)
		e.WriteString(t.encode())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *codecVolatile) DecodeTo(d *binary.Decoder, rv reflect.Value) (err error) {
	out := NewVolatile()
	size, err := d.ReadUvarint()
	if err != nil {
		return err
	}

	for i := 0; i < int(size); i++ {
		k, err := d.ReadSlice()
		if err != nil {
			return nil
		}

		v, err := d.ReadSlice()
		if err != nil {
			return nil
		}

		out.data[binary.ToString(&k)] = decodeValue(binary.ToString(&v))
	}

	rv.Set(reflect.ValueOf(*out))
	return
}
