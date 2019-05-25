// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

// This is a fork of bytes.Reader, originally licensed under
// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary

import (
	"io"
)

// A Reader implements the io.Reader, io.ReaderAt, io.WriterTo, io.Seeker,
// io.ByteScanner, and io.RuneScanner interfaces by reading from
// a byte slice.
// Unlike a Buffer, a Reader is read-only and supports seeking.
type reader struct {
	s []byte
	i int64 // current reading index
}

// Len returns the number of bytes of the unread portion of the
// slice.
func (r *reader) Len() int {
	if r.i >= int64(len(r.s)) {
		return 0
	}
	return int(int64(len(r.s)) - r.i)
}

// Size returns the original length of the underlying byte slice.
// Size is the number of bytes available for reading via ReadAt.
// The returned value is always the same and is not affected by calls
// to any other method.
func (r *reader) Size() int64 { return int64(len(r.s)) }

// Read implements the io.Reader interface.
func (r *reader) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}

	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}

// ReadByte implements the io.ByteReader interface.
func (r *reader) ReadByte() (byte, error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}

	b := r.s[r.i]
	r.i++
	return b, nil
}

// Slice selects a sub-slice of next bytes. This is similar to Read() but does not
// actually perform a copy, but simply uses the underlying slice (if available) and
// returns a sub-slice pointing to the same array. Since this requires access
// to the underlying data, this is only available for our default reader.
func (r *reader) Slice(n int) ([]byte, error) {
	if r.i+int64(n) > int64(len(r.s)) {
		return nil, io.EOF
	}

	cur := r.i
	r.i += int64(n)
	return r.s[cur:r.i], nil
}

// Reset resets the Reader to be reading from b.
func (r *reader) Reset(b []byte) {
	r.s = b
	r.i = 0
}

// newReader returns a new Reader reading from b.
func newReader(b []byte) *reader {
	return &reader{b, 0}
}
