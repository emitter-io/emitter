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
	"bytes"
	"reflect"
	"sync"

	"github.com/kelindar/binary"
)

// Reusable long-lived encoder pool.
var encoders = &sync.Pool{New: func() interface{} {
	return binary.NewEncoder(
		bytes.NewBuffer(make([]byte, 0, 8*1024)),
	)
}}

type messageCodec struct{}

// Encode encodes a value into the encoder.
func (c *messageCodec) EncodeTo(e *binary.Encoder, rv reflect.Value) (err error) {
	id := rv.Field(0).Bytes()
	channel := rv.Field(1).Bytes()
	payload := rv.Field(2).Bytes()
	ttl := rv.Field(3).Uint()

	e.WriteUvarint(uint64(len(id)))
	e.Write(id)
	e.WriteUvarint(uint64(len(channel)))
	e.Write(channel)
	e.WriteUvarint(uint64(len(payload)))
	e.Write(payload)
	e.WriteUvarint(ttl)
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *messageCodec) DecodeTo(d *binary.Decoder, rv reflect.Value) (err error) {
	var v Message
	if v.ID, err = readBytes(d); err == nil {
		if v.Channel, err = readBytes(d); err == nil {
			if v.Payload, err = readBytes(d); err == nil {
				if ttl, err := d.ReadUvarint(); err == nil {
					v.TTL = uint32(ttl)
					rv.Set(reflect.ValueOf(v))
					return nil
				}
			}
		}
	}
	return
}

func readBytes(d *binary.Decoder) (buffer []byte, err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		buffer, err = d.Slice(int(l))
	}
	return
}
