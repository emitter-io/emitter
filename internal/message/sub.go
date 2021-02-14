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

package message

import (
	"encoding/binary"
	"encoding/hex"
	"sync"
	"time"
	"unsafe"

	"github.com/emitter-io/emitter/internal/security/hash"
)

// Various constant parts of the SSID.
const (
	system        = uint32(0)
	presence      = uint32(3869262148)
	query         = uint32(3939663052)
	wildcard      = uint32(1815237614) // +
	multiWildcard = uint32(4285801373) // #
	share         = uint32(1480642916)
)

// Query represents a constant SSID for a query.
var Query = Ssid{system, query}

// Ssid represents a subscription ID which contains a contract and a list of hashes
// for various parts of the channel.
type Ssid []uint32

// NewSsid creates a new SSID.
func NewSsid(contract uint32, query []uint32) Ssid {
	ssid := make([]uint32, 0, len(query)+1)
	ssid = append(ssid, uint32(contract))
	ssid = append(ssid, query...)
	return ssid
}

// NewSsidForPresence creates a new SSID for presence.
func NewSsidForPresence(original Ssid) Ssid {
	ssid := make([]uint32, 0, len(original)+2)
	ssid = append(ssid, system)
	ssid = append(ssid, presence)
	ssid = append(ssid, original...)
	return ssid
}

// NewSsidForShare creates a new SSID for shared subscriptions.
func NewSsidForShare(original Ssid) Ssid {
	ssid := make([]uint32, 0, len(original)+1)
	ssid = append(ssid, original[0])
	ssid = append(ssid, share)
	ssid = append(ssid, original[1:]...)
	return ssid
}

// Contract gets the contract part from SSID.
func (s Ssid) Contract() uint32 {
	return uint32(s[0])
}

// GetHashCode combines the SSID into a single hash.
func (s Ssid) GetHashCode() uint32 {
	h := s[0]
	for _, i := range s[1:] {
		h ^= i
	}
	return h
}

// Encode encodes the SSID to a binary format
func (s Ssid) Encode() string {
	bin := make([]byte, 4)
	out := make([]byte, len(s)*8)

	for i, v := range s {
		if v != wildcard && v != multiWildcard {
			binary.BigEndian.PutUint32(bin, v)
			hex.Encode(out[i*8:i*8+8], bin)
		} else {
			// If we have a wildcard specified, use dot '.' symbol so this becomes a valid
			// regular expression and we could also use this for querying.
			copy(out[i*8:i*8+8], []byte{0x2E, 0x2E, 0x2E, 0x2E, 0x2E, 0x2E, 0x2E, 0x2E})
		}
	}
	return unsafeToString(out)
}

// unsafeToString is used when you really want to convert a slice
// of bytes to a string without incurring overhead. It is only safe
// to use if you really know the byte slice is not going to change
// in the lifetime of the string
func unsafeToString(bs []byte) string {
	// This is copied from runtime. It relies on the string
	// header being a prefix of the slice header!
	return *(*string)(unsafe.Pointer(&bs))
}

// ------------------------------------------------------------------------------------

// Awaiter represents an asynchronously awaiting response channel.
type Awaiter interface {
	Gather(time.Duration) [][]byte
}

// ------------------------------------------------------------------------------------

// SubscriberType represents a type of subscriber
type SubscriberType uint8

// Subscriber types
const (
	SubscriberDirect = SubscriberType(iota)
	SubscriberRemote
	SubscriberOffline
)

// Subscriber is a value associated with a subscription.
type Subscriber interface {
	ID() string
	Type() SubscriberType
	Send(*Message) error
}

// ------------------------------------------------------------------------------------

// Subscribers represents a subscriber set which can contain only unique values.
type Subscribers map[uint32]Subscriber

// NewSubscribers creates a new set of subscribers.
func newSubscribers() Subscribers {
	return make(Subscribers, 16)
}

// AddUnique adds a subscriber to the set.
func (s *Subscribers) AddUnique(value Subscriber) bool {
	if value != nil {
		key := hash.OfString(value.ID())
		if _, found := (*s)[key]; !found {
			(*s)[key] = value
			return true
		}
	}
	return false
}

// AddRange adds multiple subscribers from an existing list of subscribers, with filter applied.
func (s *Subscribers) AddRange(from Subscribers, filter func(s Subscriber) bool) {
	for id, v := range from {
		if filter == nil || filter(v) {
			(*s)[id] = v // This would simply overwrite duplicates
		}
	}
}

// Remove removes a subscriber from the set.
func (s *Subscribers) Remove(value Subscriber) bool {
	if value != nil {
		key := hash.OfString(value.ID())
		if _, ok := (*s)[key]; ok {
			delete(*s, key)
			return true
		}
	}

	return false
}

// Reset recycles the list of subscribers.
func (s *Subscribers) Reset() {
	for k := range *s {
		delete(*s, k)
	}
}

// Size returns the size of the subscriber list.
func (s *Subscribers) Size() int {
	return len(*s)
}

// Random picks a random subscriber from the map. The 'rnd' argument must be a 32-bit randomly
// generated unsigned integer in range of [0, math.MaxUint32).
func (s *Subscribers) Random(rnd uint32) (v Subscriber) {
	i, x := uint32(0), uint32((uint64(rnd)*uint64(s.Size()))>>32)
	for _, v = range *s {
		if i == x {
			break
		}
		i++
	}
	return
}

// Contains checks whether a subscriber is in the set.
func (s *Subscribers) Contains(value Subscriber) (ok bool) {
	key := hash.OfString(value.ID())
	_, ok = (*s)[key]
	return
}

// Subscription represents a topic subscription.
type Subscription struct {
	Ssid       Ssid       // Gets or sets the SSID (parsed channel) for this subscription.
	Subscriber Subscriber // Gets or sets the subscriber for this subscription.
}

// ------------------------------------------------------------------------------------

// Counters represents a subscription counting map.
type Counters struct {
	sync.Mutex
	m map[uint32]*Counter
}

// Counter represents a single subscription counter.
type Counter struct {
	Ssid    Ssid
	Channel []byte
	Counter int
}

// NewCounters creates a new container.
func NewCounters() *Counters {
	return &Counters{
		m: make(map[uint32]*Counter),
	}
}

// Increment increments the subscription counter.
func (s *Counters) Increment(ssid Ssid, channel []byte) (first bool) {
	s.Lock()
	defer s.Unlock()

	m := s.getOrCreate(ssid, channel)
	m.Counter++
	return m.Counter == 1
}

// Decrement decrements a subscription counter.
func (s *Counters) Decrement(ssid Ssid) (last bool) {
	s.Lock()
	defer s.Unlock()

	key := ssid.GetHashCode()
	if m, exists := s.m[key]; exists {
		m.Counter--

		// Remove if there's no subscribers left
		if m.Counter <= 0 {
			delete(s.m, ssid.GetHashCode())
			return true
		}
	}

	return false
}

// All returns all counters.
func (s *Counters) All() []Counter {
	s.Lock()
	defer s.Unlock()

	clone := make([]Counter, 0, len(s.m))
	for _, m := range s.m {
		clone = append(clone, *m)
	}

	return clone
}

// getOrCreate retrieves a single subscription meter or creates a new one.
func (s *Counters) getOrCreate(ssid Ssid, channel []byte) (meter *Counter) {
	key := ssid.GetHashCode()
	if m, exists := s.m[key]; exists {
		return m
	}

	meter = &Counter{
		Ssid:    ssid,
		Channel: channel,
		Counter: 0,
	}
	s.m[key] = meter
	return
}
