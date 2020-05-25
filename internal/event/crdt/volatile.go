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
	lock *sync.Mutex     // The associated mutex
	data map[string]Time // The data containing the set
}

// NewVolatile creates a new last-write-wins set with bias for 'add'.
func NewVolatile() *Volatile {
	return newVolatileWith(make(map[string]Time, 64))
}

// newVolatileWith creates a new last-write-wins set with bias for 'add'.
func newVolatileWith(items map[string]Time) *Volatile {
	return &Volatile{
		lock: new(sync.Mutex),
		data: items,
	}
}

// Add adds a value to the set.
func (s *Volatile) Add(item string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, _ := s.data[item]
	s.data[item] = Time{AddTime: Now(), DelTime: v.DelTime}
}

// Remove removes the value from the set.
func (s *Volatile) Remove(item string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, _ := s.data[item]
	s.data[item] = Time{AddTime: v.AddTime, DelTime: Now()}
}

// Contains checks if a value is present in the set.
func (s *Volatile) Contains(item string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, _ := s.data[item]
	return v.IsAdded()
}

// Get retrieves the time for an item.
func (s *Volatile) Get(item string) Time {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, _ := s.data[item]
	return v
}

// Merge merges two LWW sets. This also modifies the set being merged in
// to leave only the delta.
func (s *Volatile) Merge(other Set) {
	r := other.(*Volatile)
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
func (s *Volatile) Range(prefix []byte, f func(string, Time) bool) {
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
		v := t.Encode()

		e.WriteUvarint(uint64(len(k)))
		e.Write(stringToBinary(k))
		e.WriteUvarint(uint64(len(v)))
		e.Write(stringToBinary(v))
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
		k, err := readBytes(d)
		if err != nil {
			return nil
		}

		v, err := readBytes(d)
		if err != nil {
			return nil
		}

		out.data[binaryToString(&k)] = decodeTime(binaryToString(&v))
	}

	rv.Set(reflect.ValueOf(*out))
	return
}
