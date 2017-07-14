package encoding

import (
	"bytes"
	"io"

	"github.com/hashicorp/go-msgpack/codec"
)

// Encode serializes the content and writes it to a byte array.
func Encode(kind uint8, content interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := EncodeTo(buf, kind, content)
	return buf.Bytes(), err
}

// Decode deserializes the content from a byte array.
func Decode(buf []byte, out interface{}) error {
	return DecodeFrom(bytes.NewReader(buf), out)
}

// EncodeTo encodes the content and writes it to a writer.
func EncodeTo(writer io.Writer, kind uint8, content interface{}) error {
	writer.Write([]byte{kind})
	handle := codec.MsgpackHandle{}
	encoder := codec.NewEncoder(writer, &handle)
	err := encoder.Encode(content)
	return err
}

// DecodeFrom deserializes the content from a reader.
func DecodeFrom(reader io.Reader, out interface{}) error {
	var handle codec.MsgpackHandle
	return codec.NewDecoder(reader, &handle).Decode(out)
}
