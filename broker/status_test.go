package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getStatus(t *testing.T) {
	s := &Service{}

	status, err := s.getStatus()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, status.MemoryPrivate)
}
