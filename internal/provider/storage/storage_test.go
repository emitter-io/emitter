/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/
package storage

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
)

type survey func(string, []byte) (message.Awaiter, error)

func (s survey) Query(q string, b []byte) (message.Awaiter, error) {
	return s(q, b)
}

func testMessage(a, b, c uint32) *message.Message {
	return &message.Message{
		ID:      message.NewID(message.Ssid{0, a, b, c}),
		Channel: []byte("test/channel/"),
		Payload: []byte(fmt.Sprintf("%v,%v,%v", a, b, c)),
		TTL:     100,
	}
}

func TestNoop_Store(t *testing.T) {
	s := NewNoop()
	err := s.Store(testMessage(1, 2, 3))
	assert.NoError(t, err)
}

func TestNoop_Query(t *testing.T) {
	s := new(Noop)
	zero := time.Unix(0, 0)
	r, err := s.Query(testMessage(1, 2, 3).Ssid(), zero, zero, 10)
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

func testOrder(t *testing.T, store Storage) {
	for i := int64(0); i < 100; i++ {
		msg := message.New(message.Ssid{0, 1, 2}, []byte("a/b/c/"), []byte(fmt.Sprintf("%d", i)))
		msg.ID.SetTime(msg.ID.Time() + (i * 10000))
		assert.NoError(t, store.Store(msg))
	}

	// Issue a query
	zero := time.Unix(0, 0)
	f, err := store.Query([]uint32{0, 1, 2}, zero, zero, 5)
	assert.NoError(t, err)

	assert.Len(t, f, 5)
	assert.Equal(t, message.Ssid{0, 1, 2}, f[0].Ssid())
	assert.Equal(t, "95", string(f[0].Payload))
	assert.Equal(t, "96", string(f[1].Payload))
	assert.Equal(t, "97", string(f[2].Payload))
	assert.Equal(t, "98", string(f[3].Payload))
	assert.Equal(t, "99", string(f[4].Payload))
}

func testRetained(t *testing.T, store Storage) {

	for i := int64(0); i < 10; i++ {
		msg := message.New(message.Ssid{0, 1, 2}, []byte("a/b/c/"), []byte(fmt.Sprintf("%d", i)))
		msg.TTL = message.RetainedTTL
		msg.ID.SetTime(msg.ID.Time() + (i * 10000))
		assert.NoError(t, store.Store(msg))
	}

	// Issue a query
	zero := time.Unix(0, 0)
	f, err := store.Query([]uint32{0, 1, 2}, zero, zero, 1)
	assert.NoError(t, err)

	assert.Len(t, f, 1)
	assert.Equal(t, "9", string(f[0].Payload))
}

func testRange(t *testing.T, store Storage) {
	var t0, t1 int64
	for i := int64(0); i < 100; i++ {
		msg := message.New(message.Ssid{0, 1, 2}, []byte("a/b/c/"), []byte(fmt.Sprintf("%d", i)))
		msg.ID.SetTime(msg.ID.Time() + (i * 10000))
		if i == 50 {
			t0 = msg.ID.Time()
		}
		if i == 60 {
			t1 = msg.ID.Time()
		}

		assert.NoError(t, store.Store(msg))
	}

	// Issue a query
	f, err := store.Query([]uint32{0, 1, 2}, time.Unix(t0, 0), time.Unix(t1, 0), 5)
	assert.NoError(t, err)

	assert.Len(t, f, 5)
	assert.Equal(t, message.Ssid{0, 1, 2}, f[0].Ssid())
	assert.Equal(t, "56", string(f[0].Payload))
	assert.Equal(t, "57", string(f[1].Payload))
	assert.Equal(t, "58", string(f[2].Payload))
	assert.Equal(t, "59", string(f[3].Payload))
	assert.Equal(t, "60", string(f[4].Payload))
}

func Test_configUint32(t *testing.T) {
	raw := `{
		"provider": "memory",
		"config": {
			"retain": 99999999
		}
	}`
	cfg := testStorageConfig{}
	json.Unmarshal([]byte(raw), &cfg)

	v := configUint32(cfg.Config, "retain", 0)
	assert.Equal(t, uint32(99999999), v)
}
