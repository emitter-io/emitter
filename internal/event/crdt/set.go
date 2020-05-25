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
	bin "encoding/binary"
	"reflect"
	"time"
	"unsafe"

	"github.com/kelindar/binary"
)

// Set represents a contract for a CRDT set.
type Set interface {
	Add(string)
	Remove(string)
	Contains(string) bool
	Get(item string) Time
	Merge(Set)
	Range([]byte, func(string, Time) bool)
	Count() int
}

// New creates a new CRDT set.
func New(durable bool) Set {
	if durable {
		return NewDurable()
	}
	return NewVolatile()
}

// ------------------------------------------------------------------------------------

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

// Encode encodes the value to a string
func (t Time) Encode() string {
	b := make([]byte, 20)
	n1 := bin.PutVarint(b, t.AddTime)
	n2 := bin.PutVarint(b[n1:], t.DelTime)
	b = b[:n1+n2]
	return binaryToString(&b)
}

// DecodeTime decodes the time from a string
func decodeTime(t string) (v Time) {
	b, n := stringToBinary(t), 0
	v.AddTime, n = bin.Varint(b)
	v.DelTime, _ = bin.Varint(b[n:])
	return
}

// ------------------------------------------------------------------------------------

func readBytes(d *binary.Decoder) (buffer []byte, err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		buffer, err = d.Slice(int(l))
	}
	return
}

func binaryToString(b *[]byte) string {
	return *(*string)(unsafe.Pointer(b))
}

func stringToBinary(v string) (b []byte) {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&v))
	byteHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	byteHeader.Data = strHeader.Data

	l := len(v)
	byteHeader.Len = l
	byteHeader.Cap = l
	return
}

// ------------------------------------------------------------------------------------

// The clock for unit-testing
type clock func() int64

// Now gets the current time in Unix nanoseconds
var Now clock = func() int64 {
	return time.Now().UnixNano()
}
