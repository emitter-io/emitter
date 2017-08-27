package broker

import (
	"testing"

	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/stretchr/testify/assert"
)

func Test_getStatus(t *testing.T) {
	s := &Service{
		subcounters: subscription.NewCounters(),
	}

	status, err := s.getStatus()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, status.MemoryPrivate)
}
