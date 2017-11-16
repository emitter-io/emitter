package cluster

import (
	"io"
	"testing"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/config"
	"github.com/stretchr/testify/assert"
)

func TestOnGossipUnicast(t *testing.T) {
	frame := message.Frame{
		{Ssid: message.Ssid{1, 2, 3}, Channel: []byte("a/b/c/"), Payload: []byte("hello abc")},
		{Ssid: message.Ssid{1, 2, 3}, Channel: []byte("a/b/"), Payload: []byte("hello ab")},
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

func TestNewSwarm(t *testing.T) {
	cfg := config.ClusterConfig{
		NodeName:      "00:00:00:00:00:01",
		ListenAddr:    ":4000",
		AdvertiseAddr: ":4001",
	}

	// Create a new swarm and check if it was constructed well
	s := NewSwarm(&cfg, make(chan bool))
	assert.Equal(t, 0, s.NumPeers())
	assert.Equal(t, uint64(1), s.ID())
	assert.NotNil(t, s.Gossip())

	// Gossip with empty payload should not fail
	_, err := s.OnGossip([]byte{})
	assert.NoError(t, err)

	// Gossip with invalid data
	_, err = s.OnGossip([]byte{1, 2, 3})
	assert.Equal(t, io.EOF, err)

	// Close the swarm
	err = s.Close()
	assert.NoError(t, err)
}
