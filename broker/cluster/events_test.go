package cluster

import (
	"testing"

	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/emitter-io/emitter/encoding"
	"github.com/stretchr/testify/assert"
)

func Test_decodeMessageFrame(t *testing.T) {
	frame := MessageFrame{
		Message{Ssid: subscription.Ssid{1, 2, 3}, Channel: []byte("a/b/c/"), Payload: []byte("hello abc")},
		Message{Ssid: subscription.Ssid{1, 2, 3}, Channel: []byte("a/b/"), Payload: []byte("hello ab")},
	}

	buffer, err := encoding.Encode(&frame)
	assert.NoError(t, err)

	output, err := decodeMessageFrame(buffer)
	assert.NoError(t, err)
	assert.Equal(t, frame, output)
}
