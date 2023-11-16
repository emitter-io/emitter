package storage

import (
	"fmt"
	"testing"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestXxx(t *testing.T) {
	frame := make(message.Frame, 0, 100)
	for i := int64(0); i < 100; i++ {
		msg := message.New(message.Ssid{0, 1, 2}, []byte("a/b/c/"), []byte(fmt.Sprintf("%d", i)))
		msg.ID.SetTime(msg.ID.Time() + (i * 10000))
		frame = append(frame, *msg)
	}

	sizeLimiter := NewMessageSizeLimiter(100, 50)
	sizeLimiter.Limit(&frame)

	assert.Len(t, frame, 5)
	assert.Equal(t, message.Ssid{0, 1, 2}, frame[0].Ssid())
	assert.Equal(t, "95", string(frame[0].Payload))
	assert.Equal(t, "96", string(frame[1].Payload))
	assert.Equal(t, "97", string(frame[2].Payload))
	assert.Equal(t, "98", string(frame[3].Payload))
	assert.Equal(t, "99", string(frame[4].Payload))
}
