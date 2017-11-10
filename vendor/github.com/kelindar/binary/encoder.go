package binary

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
	"sync"
	"unsafe"
)

// Reusable long-lived encoder pool.
var encoders = &sync.Pool{New: func() interface{} {
	return &Encoder{
		Order: DefaultEndian,
	}
}}

// Marshal encodes the payload into binary format.
func Marshal(v interface{}) (output []byte, err error) {
	var buffer bytes.Buffer

	// Get the encoder from the pool, reset it
	e := encoders.Get().(*Encoder)
	e.out = &buffer
	e.err = nil

	// Encode and set the buffer if successful
	if err = e.Encode(v); err == nil {
		output = buffer.Bytes()
	}

	// Put the encoder back when we're finished
	encoders.Put(e)
	return
}

// Encoder represents a binary encoder.
type Encoder struct {
	scratch [10]byte
	out     io.Writer
	err     error
	Order   binary.ByteOrder
}

// NewEncoder creates a new encoder.
func NewEncoder(out io.Writer) *Encoder {
	return &Encoder{
		Order: DefaultEndian,
		out:   out,
	}
}

// Encode encodes the value to the binary format.
func (e *Encoder) Encode(v interface{}) (err error) {

	// Scan the type (this will load from cache)
	rv := reflect.Indirect(reflect.ValueOf(v))
	var c codec
	if c, err = scan(rv.Type()); err != nil {
		return
	}

	// Encode the value
	if err = c.EncodeTo(e, rv); err == nil {
		err = e.err
	}
	return
}

// Write writes the contents of p into the buffer.
func (e *Encoder) write(p []byte) {
	if e.err == nil {
		_, e.err = e.out.Write(p)
	}
}

// Writes a variable size integer
func (e *Encoder) writeInt64(v int64) {
	n := binary.PutVarint(e.scratch[:], v)
	e.write(e.scratch[:n])
}

// Writes a variable size unsigned integer
func (e *Encoder) writeUint64(v uint64) {
	n := binary.PutUvarint(e.scratch[:], v)
	e.write(e.scratch[:n])
}

// Writes a boolean value
func (e *Encoder) writeBool(v bool) {
	e.scratch[0] = 0
	if v {
		e.scratch[0] = 1
	}
	e.write(e.scratch[:1])
}

// Writes a complex number
func (e *Encoder) writeComplex(v complex128) {

	e.err = binary.Write(e.out, e.Order, v)
}

// Writes a floating point number
func (e *Encoder) writeFloat(v float64) {
	e.err = binary.Write(e.out, e.Order, v)
}

// Writes a string
func (e *Encoder) writeString(v string) {
	e.writeUint64(uint64(len(v)))

	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&v))

	var b []byte
	byteHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	byteHeader.Data = strHeader.Data

	l := len(v)
	byteHeader.Len = l
	byteHeader.Cap = l
	e.write(b)
}
