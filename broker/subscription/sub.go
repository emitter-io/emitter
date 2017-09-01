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

package subscription

import (
	"sync"

	"github.com/emitter-io/emitter/security"
)

// Various constant parts of the SSID.
const (
	system   = uint32(0)
	presence = uint32(3869262148)
	query    = uint32(3939663052)
	wildcard = uint32(1815237614)
)

// Query represents a constant SSID for a query.
var Query = Ssid{system, query}

// Ssid represents a subscription ID which contains a contract and a list of hashes
// for various parts of the channel.
type Ssid []uint32

// NewSsid creates a new Ssid.
func NewSsid(contract uint32, c *security.Channel) Ssid {
	ssid := make([]uint32, 0, len(c.Query)+1)
	ssid = append(ssid, uint32(contract))
	ssid = append(ssid, c.Query...)
	return ssid
}

// NewSsidForPresence creates a new Ssid for presence.
func NewSsidForPresence(original Ssid) Ssid {
	ssid := make([]uint32, 0, len(original)+2)
	ssid = append(ssid, system)
	ssid = append(ssid, presence)
	ssid = append(ssid, original...)
	return ssid
}

// Contract gets the contract part from Ssid.
func (s Ssid) Contract() uint32 {
	return uint32(s[0])
}

// GetHashCode combines the ssid into a single hash.
func (s Ssid) GetHashCode() uint32 {
	h := s[0]
	for _, i := range s[1:] {
		h ^= i
	}
	return h
}

// ------------------------------------------------------------------------------------

// SubscriberType represents a type of subscriber
type SubscriberType uint8

// Subscriber types
const (
	SubscriberDirect = SubscriberType(iota)
	SubscriberRemote
)

// Subscriber is a value associated with a subscription.
type Subscriber interface {
	ID() string
	Type() SubscriberType
	Send(ssid Ssid, channel []byte, payload []byte) error
}

// ------------------------------------------------------------------------------------

// Subscribers represents a subscriber set which can contain only unique values.
type Subscribers []Subscriber

// AddUnique adds a subscriber to the set.
func (s *Subscribers) AddUnique(value Subscriber) {
	if s.Contains(value) == false {
		*s = append(*s, value)
	}
}

// Contains checks whether a subscriber is in the set.
func (s *Subscribers) Contains(value Subscriber) bool {
	for _, v := range *s {
		if v == value {
			return true
		}
	}
	return false
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
