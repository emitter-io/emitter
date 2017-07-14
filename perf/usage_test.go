package perf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUsageTracker(t *testing.T) {
	box := NewCounters()
	net := box.NewUsageTracker("contractID")

	assert.NotNil(t, net)
	net.MessagesIn.Increment()
	assert.Equal(t, int64(1), net.MessagesIn.Value())

	net.Reset()
	assert.Equal(t, int64(0), net.MessagesIn.Value())
}
