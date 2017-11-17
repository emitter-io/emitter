package storage

import (
	"fmt"
	"testing"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/stretchr/testify/assert"
)

func testMessage(a, b, c uint32) *message.Message {
	return &message.Message{
		Time:    1354678,
		Ssid:    []uint32{0, a, b, c},
		Channel: []byte("test/channel/"),
		Payload: []byte(fmt.Sprintf("%v,%v,%v", a, b, c)),
		TTL:     10,
	}
}

func TestNoop_Store(t *testing.T) {
	s := new(Noop)
	err := s.Store(testMessage(1, 2, 3))
	assert.NoError(t, err)
}

func TestNoop_QueryLast(t *testing.T) {
	s := new(Noop)
	r, err := s.QueryLast(testMessage(1, 2, 3).Ssid, 10)
	assert.NoError(t, err)
	for range r {
		t.Errorf("Should be empty")
	}
}

func TestNoop_Configure(t *testing.T) {
	s := new(Noop)
	err := s.Configure(nil)
	assert.NoError(t, err)
}

func TestNoop_Name(t *testing.T) {
	s := new(Noop)
	assert.Equal(t, "noop", s.Name())
}

func TestNoop_Close(t *testing.T) {
	s := new(Noop)
	err := s.Close()
	assert.NoError(t, err)
}
