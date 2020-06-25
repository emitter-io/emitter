/**********************************************************************************
* Copyright (c) 2009-2020 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/

package cluster

import (
	"testing"

	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/mesh"
)

func newTestMessage(ssid message.Ssid, channel, payload string) message.Message {
	return message.Message{
		ID:      message.NewID(ssid),
		Channel: []byte(channel),
		Payload: []byte(payload),
	}
}

func TestOnGossipUnicast(t *testing.T) {
	frame := message.Frame{
		newTestMessage(message.Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(message.Ssid{1, 2, 3}, "a/b/", "hello ab"),
	}

	// Encode using binary + snappy
	encoded := frame.Encode()

	// Create a dummy swarm
	var count int
	swarm := Swarm{
		OnMessage: func(m *message.Message) {
			assert.Equal(t, frame[count], *m)
			count++
		},
	}

	// Test the unicast receive
	err := swarm.OnGossipUnicast(1, encoded)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestNewSwarm_Scenario(t *testing.T) {
	msg := newTestMessage(message.Ssid{1, 2, 3}, "a/b/c/", "hello abc")
	cfg := config.ClusterConfig{
		NodeName:      "00:00:00:00:00:01",
		ListenAddr:    ":4000",
		AdvertiseAddr: ":4001",
	}

	// Create a new swarm and check if it was constructed well
	s := NewSwarm(&cfg)
	s.update()

	assert.Equal(t, 0, s.NumPeers())
	assert.Equal(t, uint64(1), s.ID())
	assert.NotNil(t, s.Gossip())

	// Gossip with empty payload should not fail
	_, err := s.OnGossip([]byte{})
	assert.NoError(t, err)

	// Gossip with invalid data
	_, err = s.OnGossip([]byte{1, 2, 3})
	assert.Error(t, err)

	// Self broadcast should not fail
	_, err = s.OnGossipBroadcast(1, []byte{1, 2, 3})
	assert.NoError(t, err)

	// Broadcast with invalid data
	_, err = s.OnGossipBroadcast(2, []byte{1, 2, 3})
	assert.Error(t, err)

	// Find peer
	peer := s.findPeer(123)
	assert.NotNil(t, peer)

	// Send to active peer
	err = s.SendTo(123, &msg)
	assert.NoError(t, err)

	// Send to inactive peer
	peer.activity = 0
	err = s.SendTo(123, &msg)
	assert.Error(t, err)

	// Remove that peer, it should not be there
	s.onPeerOffline(123)
	assert.False(t, s.members.Contains(mesh.PeerName(123)))

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
	s := NewSwarm(&cfg)
	defer s.Close()

	// TODO: Test actual correctness as well
	assert.NotPanics(t, func() {
		s.Notify(&event.Subscription{
			Conn: 5,
			Ssid: []uint32{1, 2, 3},
		}, true)
		s.Notify(&event.Subscription{
			Conn: 5,
			Ssid: []uint32{1, 2, 3},
		}, false)
	})
}

func Test_merge(t *testing.T) {
	cfg := config.ClusterConfig{
		NodeName:      "00:00:00:00:00:01",
		ListenAddr:    ":4000",
		AdvertiseAddr: ":4001",
	}

	ev1 := &event.Subscription{
		Ssid: []uint32{1, 2, 3},
		Peer: 2,
		Conn: 30,
	}

	in := event.NewState("")
	in.Add(ev1)

	// Counter of events
	var subscribed bool

	// Create a new swarm and check if it was constructed well
	s := NewSwarm(&cfg)
	s.OnSubscribe = func(message.Subscriber, *event.Subscription) bool {
		subscribed = true
		return true
	}
	defer s.Close()

	s.members.Touch(2)
	_, err := s.merge(in.Encode()[0])
	assert.NoError(t, err)
	assert.True(t, subscribed)
	assert.True(t, s.Contains(ev1))
}

func TestJoin(t *testing.T) {
	s := new(Swarm)

	errs := s.Join("google.com", "127.0.0.1", "127.0.0.1:4000")
	assert.Empty(t, errs)
}
