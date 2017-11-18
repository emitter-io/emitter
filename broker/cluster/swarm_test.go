package cluster

import (
	"io"
	"testing"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/config"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/mesh"
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

func TestNewSwarm_Scenario(t *testing.T) {
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

	// Self broadcast should not fail
	_, err = s.OnGossipBroadcast(1, []byte{1, 2, 3})
	assert.NoError(t, err)

	// Broadcast with invalid data
	_, err = s.OnGossipBroadcast(2, []byte{1, 2, 3})
	assert.Equal(t, io.EOF, err)

	// Find peer
	peer := s.FindPeer(123)
	assert.NotNil(t, peer)
	assert.Equal(t, "00:00:00:00:00:7b", peer.ID())
	_, ok := s.members.Load(mesh.PeerName(123))
	assert.True(t, ok)

	// Remove that peer, it should not be there anymore
	s.onPeerOffline(123)
	_, ok = s.members.Load(mesh.PeerName(123))
	assert.False(t, ok)

	// Close the swarm
	err = s.Close()
	assert.NoError(t, err)
}

func TestNotify(t *testing.T) {
	cfg := config.ClusterConfig{
		NodeName:      "00:00:00:00:00:01",
		ListenAddr:    ":4000",
		AdvertiseAddr: ":4001",
	}

	// Create a new swarm and check if it was constructed well
	s := NewSwarm(&cfg, make(chan bool))
	defer s.Close()

	// TODO: Test actual correctness as well
	assert.NotPanics(t, func() {
		s.NotifySubscribe(5, []uint32{1, 2, 3})
		s.NotifyUnsubscribe(5, []uint32{1, 2, 3})
	})
}

func Test_merge(t *testing.T) {
	cfg := config.ClusterConfig{
		NodeName:      "00:00:00:00:00:01",
		ListenAddr:    ":4000",
		AdvertiseAddr: ":4001",
	}

	ev1 := SubscriptionEvent{
		Ssid: []uint32{1, 2, 3},
		Peer: 2,
		Conn: 30,
	}

	in := newSubscriptionState()
	in.Add(ev1.Encode())

	// Counter of events
	var subscribed bool

	// Create a new swarm and check if it was constructed well
	s := NewSwarm(&cfg, make(chan bool))
	s.OnSubscribe = func(message.Ssid, message.Subscriber) bool {
		subscribed = true
		return true
	}
	defer s.Close()

	_, err := s.merge(in.Encode()[0])
	assert.NoError(t, err)
	assert.True(t, subscribed)
}

func TestJoin(t *testing.T) {
	s := new(Swarm)

	errs := s.Join("google.com", "127.0.0.1", "127.0.0.1:4000")
	assert.Empty(t, errs)
}
