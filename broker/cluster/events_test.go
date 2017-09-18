package cluster

import (
	"bytes"
	"testing"

	"github.com/emitter-io/emitter/encoding"
	"github.com/stretchr/testify/assert"
)

func Test_decodeMessageFrame(t *testing.T) {
	frame := MessageFrame{
		&Message{Channel: []byte("a/b/c/"), Payload: []byte("hello abc")},
		&Message{Channel: []byte("a/b/"), Payload: []byte("hello ab")},
	}

	buffer, err := encoding.Encode(&frame)
	assert.NoError(t, err)

	decoder := encoding.NewDecoder(bytes.NewBuffer(buffer))
	output, err := decodeMessageFrame(decoder)
	assert.NoError(t, err)
	assert.Equal(t, frame, output)
}
