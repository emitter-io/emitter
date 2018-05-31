/**********************************************************************************
* Copyright (c) 2009-2018 Misakai Ltd.
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
)

const (
	fixed = 14
)

var (
	next   uint32
	unique = newUnique()
)

func newUnique() uint32 {
	b := make([]byte, 4)
	rand.Read(b)
	return binary.BigEndian.Uint32(b)
}

// SetDefaultUnique sets the unique number to use when the machine ID is not set.
func SetDefaultUnique(unique uint32) {
	unique = unique
}

// ID represents a message ID encoded at 128bit and lexigraphically sortable
type ID []byte

// NewDefaultID creates a new message identifier for the current time and default
// unique component.
func NewDefaultID(ssid Ssid) ID {
	return NewID(ssid, unique)
}

// NewID creates a new message identifier for the current time.
func NewID(ssid Ssid, unique uint32) ID {
	id := make(ID, len(ssid)*4+fixed)
	binary.BigEndian.PutUint32(id[0:4], ssid[0])
	binary.BigEndian.PutUint32(id[4:8], ssid[1])
	binary.BigEndian.PutUint64(id[8:16], uint64(math.MaxInt64-time.Now().UnixNano()))
	binary.BigEndian.PutUint16(id[16:18], uint16(atomic.AddUint32(&next, 1)))
	binary.BigEndian.PutUint32(id[18:22], unique)
	for i, v := range ssid[2:] {
		binary.BigEndian.PutUint32(id[22+i*4:26+i*4], v)
	}
	return id
}

// NewPrefix creates a new message identifier only containing the prefix.
func NewPrefix(ssid Ssid, from int64) ID {
	id := make(ID, 16)
	binary.BigEndian.PutUint32(id[0:4], ssid[0])
	binary.BigEndian.PutUint32(id[4:8], ssid[1])
	binary.BigEndian.PutUint64(id[8:16], uint64(math.MaxInt64-from))
	return id
}

// SetTime sets the time on the ID, useful for testing.
func (id ID) SetTime(t int64) {
	binary.BigEndian.PutUint64(id[8:16], uint64(math.MaxInt64-t))
}

// Time gets the time of the key, adjusted.
func (id ID) Time() int64 {
	return math.MaxInt64 - int64(binary.BigEndian.Uint64(id[8:16]))
}

// Contract retrieves the contract from the message ID.
func (id ID) Contract() uint32 {
	return binary.BigEndian.Uint32(id[0:4])
}

// Ssid retrieves the SSID from the message ID.
func (id ID) Ssid() Ssid {
	ssid := make(Ssid, (len(id)-fixed)/4)
	ssid[0] = binary.BigEndian.Uint32(id[0:4])
	ssid[1] = binary.BigEndian.Uint32(id[4:8])
	for i := 2; i < len(ssid); i++ {
		ssid[i] = binary.BigEndian.Uint32(id[14+i*4 : 18+i*4])
	}
	return ssid
}

// HasPrefix matches the prefix with the cutoff time.
func (id ID) HasPrefix(ssid Ssid, cutoff int64) bool {

	// We need the prefix to match, but we swap the channel and contract as
	// the channel has a higher probability to differ.
	if binary.BigEndian.Uint32(id[4:8]) != ssid[1] ||
		binary.BigEndian.Uint32(id[0:4]) != ssid[0] {
		return false
	}

	// Match the cutoff time, but keep in mind that the time is reversed
	return id.Time() >= cutoff
}

// Match matches the mesage ID with SSID and time bounds.
func (id ID) Match(query Ssid, from, until int64) bool {

	// We need to make sure the query is smaller than the actual encoded
	// SSID of the message, otherwise we can just skip. We also need the
	// prefix to match, but we swap the channel and contract as channel
	// has a higher probability to differ.
	if (len(query)*4) > len(id)-fixed ||
		binary.BigEndian.Uint32(id[4:8]) != query[1] ||
		binary.BigEndian.Uint32(id[0:4]) != query[0] {
		return false
	}

	// Same thing here, we iterate backwards as per assumption that the
	// likelihood of having last element of SSID matching decreases with
	// the depth of the SSID.
	for i := len(query) - 1; i > 1; i-- {
		if query[i] != binary.BigEndian.Uint32(id[fixed+i*4:fixed+4+i*4]) && query[i] != wildcard {
			return false
		}
	}

	// Match time bounds at the end, as we assume that the storage starts seeking
	// at the appropriate end and HasPrefix is called and will stop at the cutoff.
	t := id.Time()
	return t >= from && t <= until
}
