package cluster

import (
	"testing"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/mesh"
)

type stubGossip struct{}

func (s *stubGossip) GossipBroadcast(update mesh.GossipData) {}
func (s *stubGossip) GossipUnicast(dst mesh.PeerName, msg []byte) error {
	return nil
}

func TestPeer_Multiple(t *testing.T) {
	s := new(Swarm)
	p := s.newPeer(123)
	defer p.Close()

	// Make sure we have a peer
	assert.NotNil(t, p)
	assert.Empty(t, p.frame)
	assert.NotNil(t, p.cancel)
	assert.Equal(t, "00:00:00:00:00:7b", p.ID())
	assert.Equal(t, message.SubscriberRemote, p.Type())
	assert.True(t, p.IsActive())

	// Test the counters
	assert.True(t, p.onSubscribe("A", []uint32{1, 2, 3}))
	assert.False(t, p.onSubscribe("A", []uint32{1, 2, 3}))
	assert.False(t, p.onUnsubscribe("A", []uint32{1, 2, 3}))
	assert.True(t, p.onUnsubscribe("A", []uint32{1, 2, 3}))
}

func TestPeer_Send(t *testing.T) {
	s := new(Swarm)
	p := s.newPeer(123)
	defer p.Close()

	// Attach fake sender
	p.sender = new(stubGossip)

	// Make sure we have a peer
	p.Send(&message.Message{})
	assert.Equal(t, 1, len(p.frame))

	// Flush
	p.processSendQueue()
	assert.Equal(t, 0, len(p.frame))
}
