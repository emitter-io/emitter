// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
	"sync"
)

// Reusable long-lived decoder pool.
var decoders = &sync.Pool{New: func() interface{} {
	return NewDecoder(newReader(nil))
}}

// Reader represents the interface a reader should implement.
type Reader interface {
	io.Reader
	io.ByteReader
}

// Slicer represents a reader which can slice without copying.
type Slicer interface {
	Slice(n int) ([]byte, error)
}

// Unmarshal decodes the payload from the binary format.
func Unmarshal(b []byte, v interface{}) (err error) {

	// Get the decoder from the pool, reset it
	d := decoders.Get().(*Decoder)
	d.r.(*reader).Reset(b) // Reset the reader

	// Decode and set the buffer if successful and free the decoder
	err = d.Decode(v)
	decoders.Put(d)
	return
}

// Decoder represents a binary decoder.
type Decoder struct {
	r       Reader
	s       Slicer
	scratch [10]byte
}

// NewDecoder creates a binary decoder.
func NewDecoder(r Reader) *Decoder {
	var slicer Slicer
	if s, ok := r.(Slicer); ok {
		slicer = s
	}

	return &Decoder{
		r: r,
		s: slicer,
	}
}

// Decode decodes a value by reading from the underlying io.Reader.
func (d *Decoder) Decode(v interface{}) (err error) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.CanAddr() {
		return errors.New("binary: can only Decode to pointer type")
	}

	// Scan the type (this will load from cache)
	var c Codec
	if c, err = scan(rv.Type()); err == nil {
		err = c.DecodeTo(d, rv)
	}

	return
}

// Read reads a set of bytes
func (d *Decoder) Read(b []byte) (int, error) {
	return d.r.Read(b)
}

// ReadUvarint reads a variable-length Uint64 from the buffer.
func (d *Decoder) ReadUvarint() (uint64, error) {
	return binary.ReadUvarint(d.r)
}

// ReadVarint reads a variable-length Int64 from the buffer.
func (d *Decoder) ReadVarint() (int64, error) {
	return binary.ReadVarint(d.r)
}

// ReadUint16 reads a uint16
func (d *Decoder) ReadUint16() (out uint16, err error) {
	var b []byte
	if b, err = d.sliceOrScratch(2); err == nil {
		_ = b[1] // bounds check hint to compiler
		out = (uint16(b[0]) | uint16(b[1])<<8)
	}
	return
}

// ReadUint32 reads a uint32
func (d *Decoder) ReadUint32() (out uint32, err error) {
	var b []byte
	if b, err = d.sliceOrScratch(4); err == nil {
		_ = b[3] // bounds check hint to compiler
		out = (uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24)
	}
	return
}

// ReadUint64 reads a uint64
func (d *Decoder) ReadUint64() (out uint64, err error) {
	var b []byte
	if b, err = d.sliceOrScratch(8); err == nil {
		_ = b[7] // bounds check hint to compiler
		out = (uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
			uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56)
	}
	return
}

// ReadFloat32 reads a float32
func (d *Decoder) ReadFloat32() (out float32, err error) {
	var v uint32
	if v, err = d.ReadUint32(); err == nil {
		out = math.Float32frombits(v)
	}
	return
}

// ReadFloat64 reads a float64
func (d *Decoder) ReadFloat64() (out float64, err error) {
	var v uint64
	if v, err = d.ReadUint64(); err == nil {
		out = math.Float64frombits(v)
	}
	return
}

// ReadBool reads a single boolean value from the slice.
func (d *Decoder) ReadBool() (bool, error) {
	b, err := d.r.ReadByte()
	return b == 1, err
}

// ReadComplex reads a complex64
func (d *Decoder) readComplex64() (out complex64, err error) {
	err = binary.Read(d.r, binary.LittleEndian, &out)
	return
}

// ReadComplex reads a complex128
func (d *Decoder) readComplex128() (out complex128, err error) {
	err = binary.Read(d.r, binary.LittleEndian, &out)
	return
}

// sliceOrScratch a slice or reads into as scratch buffer. This is useful for values
// which will get reallocated after this, such as ints, floats, etc.
func (d *Decoder) sliceOrScratch(n int) (buffer []byte, err error) {
	if d.s != nil {
		return d.s.Slice(n)
	}

	buffer = d.scratch[:n]
	_, err = d.r.Read(buffer)
	return
}

// Slice selects a sub-slice of next bytes. This is similar to Read() but does not
// actually perform a copy, but simply uses the underlying slice (if available) and
// returns a sub-slice pointing to the same array. Since this requires access
// to the underlying data, this is only available for our default reader.
func (d *Decoder) Slice(n int) ([]byte, error) {
	if d.s != nil {
		return d.s.Slice(n)
	}

	// If we don't have a slicer, we can just allocate and read
	buffer := make([]byte, n, n)
	if _, err := d.Read(buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}
