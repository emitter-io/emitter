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

package hash

import (
	"math/bits"

	"github.com/kelindar/binary"
)

const (
	c1_32 uint32 = 0xcc9e2d51
	c2_32 uint32 = 0x1b873593
)

// OfString returns a murmur32 hash for the string
func OfString(value string) uint32 {
	return Of(binary.ToBytes(value))
}

// Of returns a murmur32 hash for the data slice.
func Of(data []byte) uint32 {
	// Seed is set to 37, same as C# version of emitter
	var h1 uint32 = 37

	clen := uint32(len(data))
	for len(data) >= 4 {
		k1 := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		data = data[4:]

		k1 *= c1_32
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2_32

		h1 ^= k1
		h1 = bits.RotateLeft32(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}

	var k1 uint32
	switch len(data) {
	case 3:
		k1 ^= uint32(data[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(data[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(data[0])
		k1 *= c1_32
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2_32
		h1 ^= k1
	}

	h1 ^= uint32(clen)
	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return (h1 << 24) | (((h1 >> 8) << 16) & 0xFF0000) | (((h1 >> 16) << 8) & 0xFF00) | (h1 >> 24)
}
