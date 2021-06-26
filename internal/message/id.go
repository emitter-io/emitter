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
	"crypto/rand"
	"encoding/binary"
	"math"
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/internal/security"
)

const (
	fixed = 16
)

// RetainedTTL represents a TTL value to use for retained messages (max TTL).
const RetainedTTL = math.MaxUint32

var (
	next   uint32
	unique = newUnique()
	offset = int64(security.MinTime) // From 2018 until 2066
)

func newUnique() uint32 {
	b := make([]byte, 4)
	rand.Read(b)
	return binary.BigEndian.Uint32(b)
}

// ID represents a message ID encoded at 128bit and lexigraphically sortable
type ID []byte

// NewID creates a new message identifier for the current time.
func NewID(ssid Ssid) ID {
	id := make(ID, len(ssid)*4+fixed)
	now := uint32(time.Now().Unix() - offset)

	binary.BigEndian.PutUint32(id[0:4], ssid[0]^ssid[1])
	binary.BigEndian.PutUint32(id[4:8], math.MaxUint32-now)
	binary.BigEndian.PutUint32(id[8:12], math.MaxUint32-atomic.AddUint32(&next, 1)) // Reverse order
	binary.BigEndian.PutUint32(id[12:16], unique)
	for i, v := range ssid {
		binary.BigEndian.PutUint32(id[fixed+i*4:fixed+4+i*4], v)
	}

	return id
}

// NewPrefix creates a new message identifier only containing the prefix.
func NewPrefix(ssid Ssid, from int64) ID {
	id := make(ID, 8)
	binary.BigEndian.PutUint32(id[0:4], ssid[0]^ssid[1])
	binary.BigEndian.PutUint32(id[4:8], math.MaxUint32-uint32(from-offset))
	return id
}

// SetTime sets the time on the ID, useful for testing.
func (id ID) SetTime(t int64) {
	binary.BigEndian.PutUint32(id[4:8], math.MaxUint32-uint32(t-offset))
}

// Time gets the time of the key, adjusted.
func (id ID) Time() int64 {
	return int64(math.MaxUint32-binary.BigEndian.Uint32(id[4:8])) + offset
}

// Contract retrieves the contract from the message ID.
func (id ID) Contract() uint32 {
	return binary.BigEndian.Uint32(id[fixed : fixed+4])
}

// Ssid retrieves the SSID from the message ID.
func (id ID) Ssid() Ssid {
	ssid := make(Ssid, (len(id)-fixed)/4)
	for i := 0; i < len(ssid); i++ {
		ssid[i] = binary.BigEndian.Uint32(id[fixed+i*4 : fixed+4+i*4])
	}
	return ssid
}

// HasPrefix matches the prefix with the cutoff time.
func (id ID) HasPrefix(ssid Ssid, cutoff int64) bool {
	return (binary.BigEndian.Uint32(id[0:4]) == ssid[0]^ssid[1]) && id.Time() >= cutoff
}

// Match matches the mesage ID with SSID and time bounds.
func (id ID) Match(query Ssid, from, until int64) bool {
	if (len(query) * 4) > len(id)-fixed {
		return false
	}

	// Same thing here, we iterate backwards as per assumption that the
	// likelihood of having last element of SSID matching decreases with
	// the depth of the SSID.
	for i := len(query) - 1; i >= 0; i-- {
		if query[i] != binary.BigEndian.Uint32(id[fixed+i*4:fixed+4+i*4]) && query[i] != wildcard && query[i] != multiWildcard {
			return false
		}
	}

	// Match time bounds at the end, as we assume that the storage starts seeking
	// at the appropriate end and HasPrefix is called and will stop at the cutoff.
	t := id.Time()
	return t >= from && t <= until
}
