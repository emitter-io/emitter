package usage

import (
	"errors"
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

func TestHTTP_Configure(t *testing.T) {
	s := NewHTTP()
	defer close(s.done)

	{
		err := s.Configure(nil)
		assert.Error(t, errors.New("Configuration was not provided for HTTP metering provider"), err)
	}

	{
		err := s.Configure(map[string]interface{}{
			"interval": 1000.0,
			"url":      "http://localhost/test",
		})
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost/test", s.url)
		assert.NotNil(t, s.http)
	}
}
