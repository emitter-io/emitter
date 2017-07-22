package cluster

import (
	"bytes"
	"testing"

	"github.com/emitter-io/emitter/encoding"
	"github.com/stretchr/testify/assert"
)

func Test_decodeSubscriptionEvent(t *testing.T) {
	event := SubscriptionEvent{
		Ssid: []uint32{1, 2, 3, 4, 5},
		Node: "hello",
	}

	buffer, err := encoding.Encode(&event)
	assert.NoError(t, err)

	output := decodeSubscriptionEvent(buffer)
	assert.Equal(t, &event, output)
}

func Test_decodeHandshakeEvent(t *testing.T) {
	event := HandshakeEvent{
		Key:  "secret key",
		Node: "hello",
	}

	buffer, err := encoding.Encode(&event)
	assert.NoError(t, err)

	decoder := encoding.NewDecoder(bytes.NewBuffer(buffer))
	output, err := decodeHandshakeEvent(decoder)
	assert.NoError(t, err)
	assert.Equal(t, &event, output)
}

func Test_decodeMessageFrame(t *testing.T) {
	frame := MessageFrame{
		Message{Channel: []byte("a/b/c/"), Payload: []byte("hello abc")},
		Message{Channel: []byte("a/b/"), Payload: []byte("hello ab")},
	}

	buffer, err := encoding.Encode(&frame)
	assert.NoError(t, err)

	decoder := encoding.NewDecoder(bytes.NewBuffer(buffer))
	output, err := decodeMessageFrame(decoder)
	assert.NoError(t, err)
	assert.Equal(t, frame, output)
}
