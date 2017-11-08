package cluster

import (
	"testing"

	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/emitter-io/emitter/collection"
	"github.com/emitter-io/emitter/encoding"
	"github.com/stretchr/testify/assert"
)

func Test_decodeMessageFrame(t *testing.T) {
	frame := MessageFrame{
		Message{Ssid: subscription.Ssid{1, 2, 3}, Channel: []byte("a/b/c/"), Payload: []byte("hello abc")},
		Message{Ssid: subscription.Ssid{1, 2, 3}, Channel: []byte("a/b/"), Payload: []byte("hello ab")},
	}

	// Encode
	buffer, err := encoding.Encode(&frame)
	assert.NoError(t, err)

	// Decode
	output, err := decodeMessageFrame(buffer)
	assert.NoError(t, err)
	assert.Equal(t, frame, output)
}

func TestEncodeSubscriptionState(t *testing.T) {
	state := (*subscriptionState)(&collection.LWWSet{
		Set: collection.LWWState{"A": {AddTime: 10, DelTime: 50}},
	})

	// Encode
	enc := state.Encode()[0]
	assert.Equal(t, []byte{0x1, 0x1, 0x41, 0x14, 0x64}, enc)

	// Decode
	dec, err := decodeSubscriptionState(enc)
	assert.NoError(t, err)
	assert.Equal(t, state, dec)
}

func TestEncodeSubscriptionEvent(t *testing.T) {
	ev := SubscriptionEvent{
		Ssid: subscription.Ssid{1, 2, 3, 4, 5},
		Peer: 657,
		Conn: 12456,
	}

	// Encode
	enc := ev.Encode()

	// Decode
	dec, err := decodeSubscriptionEvent(enc)
	assert.NoError(t, err)
	assert.Equal(t, ev, dec)
}
