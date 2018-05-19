// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

import (
	"bytes"
	"encoding/binary"
	"sort"
)

// SortedInt64s represents a sorted int64 slice.
type SortedInt64s []int64

func (s SortedInt64s) Len() int           { return len(s) }
func (s SortedInt64s) Less(i, j int) bool { return s[i] < s[j] }
func (s SortedInt64s) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// MarshalBinary implements a special purpose sortable binary encoding.
func (s SortedInt64s) MarshalBinary() (bytes []byte, err error) {
	sort.Sort(s)
	prev := int64(0)
	temp := make([]byte, 10)
	bytes = make([]byte, 0, len(s)+2)

	for _, curr := range s {
		diff := curr - prev
		bytes = append(bytes, temp[:binary.PutVarint(temp, diff)]...)
		prev = curr
	}
	return
}

// UnmarshalBinary implements a special purpose binary decoding.
func (s *SortedInt64s) UnmarshalBinary(data []byte) error {
	read := bytes.NewReader(data)
	prev := int64(0)
	for read.Len() > 0 {
		diff, _ := binary.ReadVarint(read)
		prev = prev + diff
		*s = append(*s, prev)
	}

	return nil
}

// SortedUint64s represents a sorted uint64 slice.
type SortedUint64s []uint64

func (s SortedUint64s) Len() int           { return len(s) }
func (s SortedUint64s) Less(i, j int) bool { return s[i] < s[j] }
func (s SortedUint64s) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// MarshalBinary implements a special purpose sortable binary encoding.
func (s SortedUint64s) MarshalBinary() (bytes []byte, err error) {
	sort.Sort(s)
	prev := uint64(0)
	temp := make([]byte, 10)
	bytes = make([]byte, 0, len(s)+2)

	for _, curr := range s {
		diff := curr - prev
		bytes = append(bytes, temp[:binary.PutUvarint(temp, diff)]...)
		prev = curr
	}
	return
}

// UnmarshalBinary implements a special purpose binary decoding.
func (s *SortedUint64s) UnmarshalBinary(data []byte) error {
	read := bytes.NewReader(data)
	prev := uint64(0)
	for read.Len() > 0 {
		diff, _ := binary.ReadUvarint(read)
		prev = prev + diff
		*s = append(*s, prev)
	}

	return nil
}
