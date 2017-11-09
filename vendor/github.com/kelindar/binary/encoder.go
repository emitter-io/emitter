package binary

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
	"sync"
	"unsafe"
)

// Marshal encodes the payload into binary format.
func Marshal(v interface{}) ([]byte, error) {
	b := &bytes.Buffer{}
	encoder := borrowEncoder(b)
	defer encoder.release()

	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Encoder represents a binary encoder.
type Encoder struct {
	Order binary.ByteOrder
	out   io.Writer
	buf   []byte
	n     int
	Error error
}

// NewEncoder creates a new encoder.
func NewEncoder(out io.Writer) *Encoder {
	return &Encoder{
		Order: DefaultEndian,
		out:   out,
		buf:   make([]byte, 1024),
		n:     0,
		Error: nil,
	}
}

// GetEncoder borrows a pooled encoder.
func borrowEncoder(out io.Writer) *Encoder {
	s := encoders.Get().(*Encoder)
	s.Reset(out)
	return s
}

// Encode encodes the value to the binary format.
func (e *Encoder) Encode(v interface{}) (err error) {

	// Scan the type (this will load from cache)
	rv := reflect.Indirect(reflect.ValueOf(v))
	var c codec
	if c, err = scan(rv.Type()); err != nil {
		return
	}

	// Encode and flush the encoder
	if err = c.EncodeTo(e, rv); err == nil {
		return e.Flush()
	}
	return
}

// Reusable long-lived stream pool.
var encoders = &sync.Pool{New: func() interface{} {
	return &Encoder{
		Order: DefaultEndian,
		buf:   make([]byte, 1024),
		n:     0,
		Error: nil,
	}
}}

// Release releases the stream to the pool
func (e *Encoder) release() {
	encoders.Put(e)
}

// Reset reuse this stream instance by assign a new writer
func (e *Encoder) Reset(out io.Writer) {
	e.out = out
	e.n = 0
	e.Error = nil
}

// Available returns how many bytes are unused in the buffer.
func (e *Encoder) Available() int {
	return len(e.buf) - e.n
}

// Buffered returns the number of bytes that have been written into the current buffer.
func (e *Encoder) Buffered() int {
	return e.n
}

// Buffer if writer is nil, use this method to take the result
func (e *Encoder) Buffer() []byte {
	return e.buf[:e.n]
}

// Write writes the contents of p into the buffer.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
func (e *Encoder) Write(p []byte) (nn int, err error) {
	for len(p) > e.Available() && e.Error == nil {
		if e.out == nil {
			e.growAtLeast(len(p))
		} else {
			var n int
			if e.Buffered() == 0 {
				// Large write, empty buffer.
				// Write directly from p to avoid copy.
				n, e.Error = e.out.Write(p)
			} else {
				n = copy(e.buf[e.n:], p)
				e.n += n
				e.Flush()
			}
			nn += n
			p = p[n:]
		}
	}
	if e.Error != nil {
		return nn, e.Error
	}
	n := copy(e.buf[e.n:], p)
	e.n += n
	nn += n
	return nn, nil
}

func (e *Encoder) writeInt64(v int64) {
	if e.Error == nil {
		e.ensure(10)
		e.n += binary.PutVarint(e.buf[e.n:], v)
	}
}

func (e *Encoder) writeUint64(v uint64) {
	if e.Error == nil {
		e.ensure(10)
		e.n += binary.PutUvarint(e.buf[e.n:], v)
	}
}

func (e *Encoder) writeBool(v bool) {
	if e.Error == nil {
		e.ensure(1)

		var b byte
		if v {
			b = 1
		}
		e.buf[e.n] = b
		e.n++
	}
}

func (e *Encoder) writeString(v string) {
	e.writeUint64(uint64(len(v)))

	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&v))

	var b []byte
	byteHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	byteHeader.Data = strHeader.Data

	l := len(v)
	byteHeader.Len = l
	byteHeader.Cap = l
	e.Write(b)
}

// Flush writes any buffered data to the underlying io.Writer.
func (e *Encoder) Flush() error {
	if e.out == nil {
		return nil
	}
	if e.Error != nil {
		return e.Error
	}
	if e.n == 0 {
		return nil
	}
	n, err := e.out.Write(e.buf[0:e.n])
	if n < e.n && err == nil {
		err = io.ErrShortWrite
	}
	if err != nil {
		if n > 0 && n < e.n {
			copy(e.buf[0:e.n-n], e.buf[n:e.n])
		}
		e.n -= n
		e.Error = err
		return err
	}
	e.n = 0
	return nil
}

func (e *Encoder) ensure(minimal int) {
	available := e.Available()
	if available < minimal {
		e.growAtLeast(minimal)
	}
}

func (e *Encoder) growAtLeast(minimal int) {
	if e.out != nil {
		e.Flush()
		if e.Available() >= minimal {
			return
		}
	}
	toGrow := len(e.buf)
	if toGrow < minimal {
		toGrow = minimal
	}
	newBuf := make([]byte, len(e.buf)+toGrow)
	copy(newBuf, e.Buffer())
	e.buf = newBuf
}
