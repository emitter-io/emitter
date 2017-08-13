package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getStatus(t *testing.T) {
	s := &Service{
		subcounters: NewSubscriptionCounters(),
	}

	status := s.getStatus()
	assert.NotEqual(t, 0, status.MemoryPrivate)
}
