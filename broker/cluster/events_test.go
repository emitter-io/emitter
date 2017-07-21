package cluster

import (
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

	output := decodeHandshakeEvent(buffer)
	assert.Equal(t, &event, output)
}
