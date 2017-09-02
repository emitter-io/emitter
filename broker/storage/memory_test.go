package storage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testStorageConfig struct {
	Provider string                 `json:"provider"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

func TestInMemory_Configure(t *testing.T) {
	s := new(InMemory)
	cfg := map[string]interface{}{
		"maxsize": float64(1),
		"prune":   float64(1),
	}

	err := s.Configure(cfg)
	assert.NoError(t, err)

	errClose := s.Close()
	assert.NoError(t, errClose)
}

func TestInMemory_Store(t *testing.T) {
	s := new(InMemory)
	s.Configure(nil)

	err := s.Store([]uint32{1, 2, 3}, []byte("test"), 10*time.Second)
	assert.NoError(t, err)

	assert.Equal(t, []byte("test"), s.mem.Get("0000000100000002:1").Value().(message).payload)
}

func TestInMemory_lookup(t *testing.T) {
	s := new(InMemory)
	s.Configure(nil)
	s.Store([]uint32{0, 1, 1, 1}, []byte("1,1,1"), 10*time.Second)
	s.Store([]uint32{0, 1, 1, 2}, []byte("1,1,2"), 10*time.Second)
	s.Store([]uint32{0, 1, 2, 1}, []byte("1,2,1"), 10*time.Second)
	s.Store([]uint32{0, 1, 2, 2}, []byte("1,2,2"), 10*time.Second)
	s.Store([]uint32{0, 1, 3, 1}, []byte("1,3,1"), 10*time.Second)
	s.Store([]uint32{0, 1, 3, 2}, []byte("1,3,2"), 10*time.Second)

	const wildcard = uint32(1815237614)
	tests := []struct {
		query []uint32
		limit int
		count int
	}{
		{query: []uint32{0, 10, 20, 50}, limit: 10, count: 0},
		{query: []uint32{0, 1, 1, 1}, limit: 10, count: 1},
		{query: []uint32{0, 1, 1, wildcard}, limit: 10, count: 2},
		{query: []uint32{0, 1}, limit: 10, count: 6},
		{query: []uint32{0, 2}, limit: 10, count: 0},
		{query: []uint32{0, 1, 2}, limit: 10, count: 2},
	}

	for _, tc := range tests {
		matches := s.lookup(tc.query, tc.limit)
		assert.Equal(t, tc.count, len(matches))
	}
}

func Test_param(t *testing.T) {
	raw := `{
	"provider": "memory",
	"config": {
		"maxsize": 99999999
	}
}`
	cfg := testStorageConfig{}
	json.Unmarshal([]byte(raw), &cfg)

	v := param(cfg.Config, "maxsize", 0)
	assert.Equal(t, int64(99999999), v)
}

func Test_messageSize(t *testing.T) {
	msg := message{ssid: "abc", payload: []byte("hello")}
	assert.Equal(t, int64(5), msg.Size())
}
