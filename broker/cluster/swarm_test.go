package cluster

import (
	"testing"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/stretchr/testify/assert"
)

func TestOnGossipUnicast(t *testing.T) {
	frame := message.Frame{
		{Ssid: subscription.Ssid{1, 2, 3}, Channel: []byte("a/b/c/"), Payload: []byte("hello abc")},
		{Ssid: subscription.Ssid{1, 2, 3}, Channel: []byte("a/b/"), Payload: []byte("hello ab")},
	}

	// Encode using binary + snappy
	encoded, err := frame.Encode()
	assert.NoError(t, err)

	// Create a dummy swarm
	var count int
	swarm := Swarm{
		OnMessage: func(m *message.Message) {
			assert.Equal(t, frame[count], *m)
			count++
		},
	}

	// Test the unicast receive
	err = swarm.OnGossipUnicast(1, encoded)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}
