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
	"time"

	"github.com/kelindar/binary"
)

// Map represents a contract for a CRDT map.
type Map interface {
	Add(string, []byte)
	Del(string)
	Has(string) bool
	Get(string) Value
	Merge(Map)
	Range([]byte, bool, func(string, Value) bool)
	Count() int
}

// New creates a new CRDT map.
func New(durable bool, path string) Map {
	if durable {
		return NewDurable(path)
	}
	return NewVolatile()
}

// ------------------------------------------------------------------------------------

// Value represents a time pair with a value.
type Value []byte

// newValue returns zero time.
func newValue() Value {
	return Value(make([]byte, 16))
}

// decodeValue decodes the time from a string
func decodeValue(t string) Value {
	return Value(binary.ToBytes(t))
}

// IsZero checks if the time is zero
func (v Value) IsZero() bool {
	return (v.AddTime() == 0 && v.DelTime() == 0)
}

// IsAdded checks if add time is larger than remove time.
func (v Value) IsAdded() bool {
	return v.AddTime() != 0 && v.AddTime() >= v.DelTime()
}

// IsRemoved checks if remove time is larger than add time.
func (v Value) IsRemoved() bool {
	return v.AddTime() < v.DelTime()
}

// AddTime returns when the entry was added.
func (v Value) AddTime() int64 {
	return int64(binary.BigEndian.Uint64(v[0:8]))
}

// setAddTime sets when the entry was added.
func (v Value) setAddTime(t int64) {
	binary.BigEndian.PutUint64(v[0:8], uint64(t))
}

// DelTime returns when the entry was removed.
func (v Value) DelTime() int64 {
	return int64(binary.BigEndian.Uint64(v[8:16]))
}

// setDelTime sets when the entry was removed.
func (v Value) setDelTime(t int64) {
	binary.BigEndian.PutUint64(v[8:16], uint64(t))
}

// Value returns the extra value payload.
func (v Value) Value() []byte {
	return v[16:]
}

// setValue sets the extra value payload.
func (v *Value) setValue(p []byte) {
	h := (*v)[:16]
	*v = append(h, p...)
}

// encode encodes the value to a string
func (v *Value) encode() string {
	return binary.ToString((*[]byte)(v))
}

// ------------------------------------------------------------------------------------

// The clock for unit-testing
type clock func() int64

// Now gets the current time in Unix nanoseconds
var Now clock = func() int64 {
	return time.Now().UnixNano()
}
