package encoding

import (
	"io"

	"github.com/kelindar/binary"
)

// Decoder represents a decoder which can read from a stream.
type Decoder interface {
	Decode(interface{}) error
}

// Encoder represents an encoder which can write to a stream.
type Encoder interface {
	Encode(interface{}) error
}

// NewDecoder constructs a new decoder
func NewDecoder(reader binary.Reader) Decoder {
	return binary.NewDecoder(reader)
}

// NewEncoder constructs a new decoder
func NewEncoder(writer io.Writer) Encoder {
	return binary.NewEncoder(writer)
}

// Encode serializes the content and writes it to a byte array.
func Encode(content interface{}) ([]byte, error) {
	return binary.Marshal(content)
}

// Decode deserializes the content from a byte array.
func Decode(buf []byte, out interface{}) error {
	return binary.Unmarshal(buf, out)
}
