package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoop_Store(t *testing.T) {
	s := new(Noop)
	err := s.Store([]uint32{1, 2, 3}, []byte("test"), 10*time.Second)
	assert.NoError(t, err)
}

func TestNoop_QueryLast(t *testing.T) {
	s := new(Noop)
	r, err := s.QueryLast([]uint32{1, 2, 3}, 10)
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

func TestNoop_Close(t *testing.T) {
	s := new(Noop)
	err := s.Close()
	assert.NoError(t, err)
}
