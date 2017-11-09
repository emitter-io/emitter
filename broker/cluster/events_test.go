package cluster

import (
	"testing"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/collection"
	"github.com/stretchr/testify/assert"
)

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
		Ssid: message.Ssid{1, 2, 3, 4, 5},
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
