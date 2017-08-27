package subscription

import (
	"testing"

	"github.com/emitter-io/emitter/security"
	"github.com/stretchr/testify/assert"
)

func TestSsid(t *testing.T) {
	c := security.Channel{
		Key:         []byte("key"),
		Channel:     []byte("a/b/c/"),
		Query:       []uint32{10, 20, 50},
		Options:     []security.ChannelOption{},
		ChannelType: security.ChannelStatic,
	}

	ssid := NewSsid(0, &c)
	assert.Equal(t, uint32(0), ssid.Contract())
	assert.Equal(t, uint32(0x2c), ssid.GetHashCode())
}
