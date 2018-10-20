// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package sorted

import (
	"reflect"

	"github.com/kelindar/binary"
)

// ------------------------------------------------------------------------------

// Uint16s represents a slice serialized in an unsafe, non portable manner.
type Uint16s []uint16

func (s Uint16s) Len() int           { return len(s) }
func (s Uint16s) Less(i, j int) bool { return s[i] < s[j] }
func (s Uint16s) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// GetBinaryCodec retrieves a custom binary codec.
func (s *Uint16s) GetBinaryCodec() binary.Codec {
	return UintsCodecAs(reflect.TypeOf(Uint16s{}), 2)
}

// ------------------------------------------------------------------------------

// Int16s represents a slice serialized in an unsafe, non portable manner.
type Int16s []int16

func (s Int16s) Len() int           { return len(s) }
func (s Int16s) Less(i, j int) bool { return s[i] < s[j] }
func (s Int16s) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// GetBinaryCodec retrieves a custom binary codec.
func (s *Int16s) GetBinaryCodec() binary.Codec {
	return IntsCodecAs(reflect.TypeOf(Int16s{}), 2)
}

// ------------------------------------------------------------------------------

// Uint32s represents a slice serialized in an unsafe, non portable manner.
type Uint32s []uint32

func (s Uint32s) Len() int           { return len(s) }
func (s Uint32s) Less(i, j int) bool { return s[i] < s[j] }
func (s Uint32s) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// GetBinaryCodec retrieves a custom binary codec.
func (s *Uint32s) GetBinaryCodec() binary.Codec {
	return UintsCodecAs(reflect.TypeOf(Uint32s{}), 4)
}

// ------------------------------------------------------------------------------

// Int32s represents a slice serialized in an unsafe, non portable manner.
type Int32s []int32

func (s Int32s) Len() int           { return len(s) }
func (s Int32s) Less(i, j int) bool { return s[i] < s[j] }
func (s Int32s) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// GetBinaryCodec retrieves a custom binary codec.
func (s *Int32s) GetBinaryCodec() binary.Codec {
	return IntsCodecAs(reflect.TypeOf(Int32s{}), 4)
}

// ------------------------------------------------------------------------------

// Uint64s represents a slice serialized in an unsafe, non portable manner.
type Uint64s []uint64

func (s Uint64s) Len() int           { return len(s) }
func (s Uint64s) Less(i, j int) bool { return s[i] < s[j] }
func (s Uint64s) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// GetBinaryCodec retrieves a custom binary codec.
func (s *Uint64s) GetBinaryCodec() binary.Codec {
	return UintsCodecAs(reflect.TypeOf(Uint64s{}), 8)
}

// ------------------------------------------------------------------------------

// Int64s represents a slice serialized in an unsafe, non portable manner.
type Int64s []int64

func (s Int64s) Len() int           { return len(s) }
func (s Int64s) Less(i, j int) bool { return s[i] < s[j] }
func (s Int64s) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// GetBinaryCodec retrieves a custom binary codec.
func (s *Int64s) GetBinaryCodec() binary.Codec {
	return IntsCodecAs(reflect.TypeOf(Int64s{}), 8)
}
