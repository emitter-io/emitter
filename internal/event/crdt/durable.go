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
	"math/rand"
	"reflect"
	"sort"
	"time"

	"github.com/coocood/freecache"
	"github.com/kelindar/binary"
	"github.com/tidwall/buntdb"
)

// getTime retrieves a time from the store.
func getTime(tx *buntdb.Tx, item string) Time {
	if t, err := tx.Get(item); err == nil {
		return decodeTime(t)
	}
	return Time{}
}

// Durable represents a last-write-wins CRDT set which can be persisted to disk.
type Durable struct {
	db    *buntdb.DB       // The underlying data store
	cache *freecache.Cache // Cache to use for contains checks.
}

// NewDurable creates a new last-write-wins set with bias for 'add'.
func NewDurable(dir string) *Durable {
	return newDurableWith(dir, nil)
}

// newDurableWith creates a new last-write-wins set with bias for 'add'.
func newDurableWith(path string, items map[string]Time) *Durable {
	if path == "" {
		path = ":memory:"
	}

	cache := freecache.NewCache(1 << 20) // 1MB
	db, err := buntdb.Open(path)
	if err != nil {
		panic(err)
	}

	s := &Durable{
		cache: cache,
		db:    db,
	}

	for k, v := range items {
		s.db.Update(func(tx *buntdb.Tx) error {
			s.store(tx, k, v)
			return nil
		})
	}
	return s
}

// Store stores the item into the transaction.
func (s *Durable) store(tx *buntdb.Tx, key string, t Time) {
	var opts *buntdb.SetOptions
	if t.IsRemoved() {
		opts = &buntdb.SetOptions{
			Expires: true,
			TTL:     6 * time.Hour,
		}
	}

	tx.Set(key, t.Encode(), opts)
}

// Fetch fetches the item either from transaction or cache.
func (s *Durable) fetch(item string) Time {
	cacheKey := binary.ToBytes(item)
	if v, err := s.cache.Get(cacheKey); err == nil {
		return decodeTime(binary.ToString(&v))
	}

	tx, _ := s.db.Begin(false)
	defer tx.Rollback()
	if t, err := tx.Get(item); err == nil {
		s.cache.Set(cacheKey, binary.ToBytes(t), 60)
		return decodeTime(t)
	}
	return Time{}
}

// Add adds a value to the set.
func (s *Durable) Add(item string) {
	s.db.Update(func(tx *buntdb.Tx) error {
		t := getTime(tx, item)
		s.store(tx, item, Time{AddTime: Now(), DelTime: t.DelTime})
		return nil
	})
}

// Remove removes the value from the set.
func (s *Durable) Remove(item string) {
	s.db.Update(func(tx *buntdb.Tx) error {
		v := getTime(tx, item)
		s.store(tx, item, Time{AddTime: v.AddTime, DelTime: Now()})
		return nil
	})
}

// Contains checks if a value is present in the set.
func (s *Durable) Contains(item string) bool {
	return s.fetch(item).IsAdded()
}

// Get retrieves the time for an item.
func (s *Durable) Get(item string) Time {
	return s.fetch(item)
}

// Merge merges two LWW sets. This also modifies the set being merged in
// to leave only the delta.
func (s *Durable) Merge(other Set) {
	r := other.(*Volatile)
	r.lock.Lock()
	defer r.lock.Unlock()

	s.db.Update(func(stx *buntdb.Tx) error {
		for key, rt := range r.data {
			st := getTime(stx, key)

			// Update add time
			if st.AddTime < rt.AddTime {
				st.AddTime = rt.AddTime
			} else {
				rt.AddTime = 0 // Remove from delta
			}

			// Update delete time
			if st.DelTime < rt.DelTime {
				st.DelTime = rt.DelTime
			} else {
				rt.DelTime = 0 // Remove from delta
			}

			if rt.IsZero() {
				delete(r.data, key) // Remove from delta
			} else {
				s.store(stx, key, st) // Merge the new value
				r.data[key] = rt      // Update the delta
			}
		}

		return nil
	})
}

// Range iterates through the events for a specific prefix.
func (s *Durable) Range(prefix []byte, f func(string, Time) bool) {
	s.db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("", func(k, v string) bool {
			if !bytes.HasPrefix(binary.ToBytes(k), prefix) {
				return true
			}

			return f(k, decodeTime(v))
		})
	})
}

// Count returns the number of items in the set.
func (s *Durable) Count() (count int) {
	s.Range(nil, func(k string, v Time) bool {
		count++
		return true
	})
	return
}

// ToMap converts the set to a map (useful for testing).
func (s *Durable) toMap() map[string]Time {
	m := make(map[string]Time)
	s.Range(nil, func(k string, v Time) bool {
		m[k] = v
		return true
	})
	return m
}

// Close closes the set gracefully.
func (s *Durable) Close() error {
	return s.db.Close()
}

// GetBinaryCodec retrieves a custom binary codec.
func (s *Durable) GetBinaryCodec() binary.Codec {
	return new(durableCodec)
}

// ------------------------------------------------------------------------------------

type entry struct {
	Key string
	Val string
}

type durableCodec struct{}

// Encode encodes a value into the encoder.
func (c *durableCodec) EncodeTo(e *binary.Encoder, rv reflect.Value) (err error) {
	s := rv.Interface().(Durable)

	count, reservoirSize := 0, 50000
	entries := make([]entry, 0, reservoirSize)

	s.db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("", func(k, v string) bool {
			if count++; count <= reservoirSize {
				entries = append(entries, entry{
					Key: k,
					Val: v,
				})
				return true
			}

			// Try to randomly substitute
			if r := int(rand.Int31n(int32(count))); r < reservoirSize {
				entries[r] = entry{
					Key: k,
					Val: v,
				}
			}
			return true
		})
	})

	// Sort the slice by key
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })

	// Write the entire sample
	e.WriteUvarint(uint64(len(entries)))
	for _, v := range entries {
		e.WriteString(v.Key)
		e.WriteString(v.Val)
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *durableCodec) DecodeTo(d *binary.Decoder, rv reflect.Value) (err error) {
	out := NewDurable("")
	size, err := d.ReadUvarint()
	if err != nil {
		return err
	}

	out.db.Update(func(tx *buntdb.Tx) error {
		for i := 0; i < int(size); i++ {
			k, err := d.ReadSlice()
			if err != nil {
				return nil
			}

			v, err := d.ReadSlice()
			if err != nil {
				return nil
			}

			tx.Set(binary.ToString(&k), binary.ToString(&v), nil)
		}
		return nil
	})

	rv.Set(reflect.ValueOf(*out))
	return
}
