package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_peerKey(t *testing.T) {
	node := "Peer1"
	expected := uint32(0x6e682264)

	assert.Equal(t, expected, peerKey(node))
}
