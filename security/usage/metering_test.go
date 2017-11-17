package usage

import (
	"errors"
	"testing"

	"github.com/emitter-io/emitter/network/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestHTTP_Store(t *testing.T) {
	h := http.NewMockClient()
	h.On("PostJSON", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	u1 := usage{MessageIn: 1, TrafficIn: 200, MessageEg: 1, TrafficEg: 100, Contract: 0x1}
	u2 := usage{MessageIn: 0, TrafficIn: 0, MessageEg: 0, TrafficEg: 0, Contract: 0x1}

	s := NewHTTP()
	s.http = h

	c := s.Get(1)
	c.AddEgress(100)
	c.AddIngress(200)
	assert.Equal(t, &u1, c)

	s.store()
	assert.Equal(t, &u2, c)
}
