package usage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoop_New(t *testing.T) {
	s := NewNoop()
	assert.Equal(t, &NoopStorage{}, s)
}

func TestNoop_Configure(t *testing.T) {
	s := new(NoopStorage)
	err := s.Configure(nil)
	assert.NoError(t, err)
}

func TestNoop_Name(t *testing.T) {
	s := new(NoopStorage)
	assert.Equal(t, "noop", s.Name())
}

func TestNoop_Get(t *testing.T) {
	s := new(NoopStorage)
	assert.Equal(t, uint32(123), s.Get(123).(Meter).GetContract())
}

func TestHTTP_New(t *testing.T) {
	s := NewHTTP()
	assert.NotNil(t, s.counters)
}

func TestHTTP_Name(t *testing.T) {
	s := new(HTTPStorage)
	assert.Equal(t, "http", s.Name())
}
