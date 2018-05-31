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
	"encoding/binary"
	"math"
	"time"
)

// Key represents an efficient binary-encoded key for message storage.
type Key []byte

// NewPrefix creates a new prefix key which can be used for a SEEK operation.
func NewPrefix(ssid Ssid, from int64) Key {
	out := make(Key, 16)
	binary.BigEndian.PutUint32(out[0:4], ssid[0])
	binary.BigEndian.PutUint32(out[4:8], ssid[1])
	binary.BigEndian.PutUint64(out[8:16], uint64(math.MaxInt64-from))
	return out
}

// Key encodes the key for storage and matching.
func (m *Message) Key() Key {
	if m.Time == 0 {
		m.Time = time.Now().UnixNano()
	}

	// Encode the key
	out := make(Key, len(m.Ssid)*4+8) //+8)
	binary.BigEndian.PutUint32(out[0:4], m.Ssid[0])
	binary.BigEndian.PutUint32(out[4:8], m.Ssid[1])
	binary.BigEndian.PutUint64(out[8:16], uint64(math.MaxInt64-m.Time))
	for i, v := range m.Ssid[2:] {
		binary.BigEndian.PutUint32(out[16+i*4:20+i*4], v)
	}
	return out
}

// HasPrefix matches the prefix with the cutoff time.
func (b Key) HasPrefix(query Ssid, cutoff int64) bool {

	// We need the prefix to match, but we swap the channel and contract as
	// the channel has a higher probability to differ.
	if (len(query)*4) > len(b)-8 ||
		binary.BigEndian.Uint32(b[4:8]) != query[1] ||
		binary.BigEndian.Uint32(b[0:4]) != query[0] {
		return false
	}

	// Match the cutoff time, but keep in mind that the time is reversed
	return b.Time() >= cutoff
}

// Match matches the key with SSID and time bounds.
func (b Key) Match(query Ssid, from, until int64) bool {

	// We need to make sure the query is smaller than the actual encoded
	// SSID of the message, otherwise we can just skip. We also need the
	// prefix to match, but we swap the channel and contract as channel
	// has a higher probability to differ.
	if (len(query)*4) > len(b)-8 ||
		binary.BigEndian.Uint32(b[4:8]) != query[1] ||
		binary.BigEndian.Uint32(b[0:4]) != query[0] {
		return false
	}

	// Same thing here, we iterate backwards as per assumption that the
	// likelihood of having last element of SSID matching decreases with
	// the depth of the SSID.
	for i := len(query) - 1; i > 1; i-- {
		if query[i] != binary.BigEndian.Uint32(b[8+i*4:12+i*4]) && query[i] != wildcard {
			return false
		}
	}

	// Match time bounds at the end, as we assume that the storage starts seeking
	// at the appropriate end and HasPrefix is called and will stop at the cutoff.
	t := b.Time()
	return t >= from && t <= until
}

// Time gets the time of the key, adjusted.
func (b Key) Time() int64 {
	return math.MaxInt64 - int64(binary.BigEndian.Uint64(b[8:16]))
}
