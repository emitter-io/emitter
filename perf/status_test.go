package perf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNetworkCounters(t *testing.T) {
	box := NewCounters()
	net := box.NewNetworkCounters()

	assert.NotNil(t, net)
}
