package usage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoop_Configure(t *testing.T) {
	s := new(NoopStorage)
	err := s.Configure(nil)
	assert.NoError(t, err)
}

func TestNoop_Name(t *testing.T) {
	s := new(NoopStorage)
	assert.Equal(t, "noop", s.Name())
}

func TestNoop_Store(t *testing.T) {
	s := new(NoopStorage)
	assert.NoError(t, s.Store())
}

func TestNoop_Get(t *testing.T) {
	s := new(NoopStorage)
	assert.Equal(t, uint32(123), s.Get(123).GetContract())
}
