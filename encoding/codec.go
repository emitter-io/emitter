package encoding

import (
	"bytes"
	"io"

	"github.com/hashicorp/go-msgpack/codec"
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
func NewDecoder(reader io.Reader) Decoder {
	var handle codec.BincHandle
	return codec.NewDecoder(reader, &handle)
}

// NewEncoder constructs a new decoder
func NewEncoder(writer io.Writer) Encoder {
	var handle codec.BincHandle
	return codec.NewEncoder(writer, &handle)
}

// Encode serializes the content and writes it to a byte array.
func Encode(content interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := EncodeTo(buf, content)
	return buf.Bytes(), err
}

// Decode deserializes the content from a byte array.
func Decode(buf []byte, out interface{}) error {
	handle := codec.BincHandle{}
	return codec.NewDecoderBytes(buf, &handle).Decode(out)
}

// EncodeTo encodes the content and writes it to a writer.
func EncodeTo(writer io.Writer, content interface{}) error {
	return NewEncoder(writer).Encode(content)
}

// DecodeFrom deserializes the content from a reader.
func DecodeFrom(reader io.Reader, out interface{}) error {
	return NewDecoder(reader).Decode(out)
}
