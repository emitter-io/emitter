package utils

import (
	"github.com/kelindar/binary"
)

// Encode serializes the content and writes it to a byte array.
func Encode(content interface{}) ([]byte, error) {
	return binary.Marshal(content)
}

// Decode deserializes the content from a byte array.
func Decode(buf []byte, out interface{}) error {
	return binary.Unmarshal(buf, out)
}
