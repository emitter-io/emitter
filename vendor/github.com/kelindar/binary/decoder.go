package binary

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

// Reader represents the interface a reader should implement
type Reader interface {
	io.Reader
	io.ByteReader
}

// Unmarshal decodes the payload from the binary format.
func Unmarshal(b []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(b)).Decode(v)
}

// Decoder represents a binary decoder.
type Decoder struct {
	Order binary.ByteOrder
	r     Reader
}

// NewDecoder creates a binary decoder.
func NewDecoder(r Reader) *Decoder {
	return &Decoder{
		Order: DefaultEndian,
		r:     r,
	}
}

// Decode decodes a value by reading from the underlying io.Reader.
func (d *Decoder) Decode(v interface{}) (err error) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.CanAddr() {
		return errors.New("binary: can only Decode to pointer type")
	}

	// Scan the type (this will load from cache)
	var c codec
	if c, err = scan(rv.Type()); err == nil {
		err = c.DecodeTo(d, rv)
	}

	return
}
