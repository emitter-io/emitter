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
	"reflect"
	"unsafe"

	"github.com/kelindar/binary"
)

// GetBinaryCodec retrieves a custom binary codec.
func (s *Set) GetBinaryCodec() binary.Codec {
	return new(codec)
}

type codec struct{}

// Encode encodes a value into the encoder.
func (c *codec) EncodeTo(e *binary.Encoder, rv reflect.Value) (err error) {
	s := rv.Interface().(Set)
	s.lock.Lock()
	defer s.lock.Unlock()

	// Since we're iterating over a map, the iteration should be done in pseudo-random
	// order. Hence, we take advantage of this and break at 100K subscriptions in order
	// to make sure the gossip message fits under 10MB (max size).
	count, size := 0, 100000
	if len(s.data) < size {
		size = len(s.data)
	}

	e.WriteUvarint(uint64(size))
	for k, v := range s.data {
		e.WriteVarint(v.AddTime)
		e.WriteVarint(v.DelTime)
		e.WriteUvarint(uint64(len(k)))
		e.Write(stringToBinary(k))
		if count++; count >= size {
			break
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *codec) DecodeTo(d *binary.Decoder, rv reflect.Value) (err error) {
	out := New()
	size, err := d.ReadUvarint()
	if err != nil {
		return err
	}

	var addTime, delTime int64
	for i := 0; i < int(size); i++ {
		if addTime, err = d.ReadVarint(); err == nil {
			if delTime, err = d.ReadVarint(); err == nil {
				k, err := readBytes(d)
				if err != nil {
					return nil
				}

				out.data[binaryToString(&k)] = Time{
					AddTime: addTime,
					DelTime: delTime,
				}
			}
		}
	}

	rv.Set(reflect.ValueOf(*out))
	return
}

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
