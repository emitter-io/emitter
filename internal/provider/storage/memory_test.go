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
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
)

// awaiter represents a query awaiter.
type mockAwaiter struct {
	f func(timeout time.Duration) (r [][]byte)
}

func (a *mockAwaiter) Gather(timeout time.Duration) (r [][]byte) {
	return a.f(timeout)
}

type testStorageConfig struct {
	Provider string                 `json:"provider"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

func newTestMemStore() *InMemory {
	s := new(InMemory)
	s.Configure(nil)

	s.Store(testMessage(1, 1, 1))
	s.Store(testMessage(1, 1, 2))
	s.Store(testMessage(1, 2, 1))
	s.Store(testMessage(1, 2, 2))
	s.Store(testMessage(1, 3, 1))
	s.Store(testMessage(1, 3, 2))
	return s
}

func TestInMemory_Name(t *testing.T) {
	s := NewInMemory(nil)
	assert.Equal(t, "inmemory", s.Name())
}

func TestInMemory_Configure(t *testing.T) {
	s := new(InMemory)
	err := s.Configure(map[string]interface{}{})
	assert.NoError(t, err)

	errClose := s.Close()
	assert.NoError(t, errClose)
}

func TestInMemory_QueryOrdered(t *testing.T) {
	store := new(InMemory)
	store.Configure(nil)
	testOrder(t, store)
}

func TestInMemory_QueryRange(t *testing.T) {
	store := new(InMemory)
	store.Configure(nil)
	testRange(t, store)
}

func TestInMemory_QueryRetained(t *testing.T) {
	store := new(InMemory)
	store.Configure(nil)
	testRetained(t, store)
}

func TestInMemory_Store(t *testing.T) {
	s := new(InMemory)
	s.Configure(nil)

	msg := testMessage(1, 2, 3)
	//msg.Time = 0
	err := s.Store(msg)
	assert.NoError(t, err)
	//assert.Equal(t, []byte("1,2,3"), s.mem.Get("0000000000000001:1").Value().(message.Message).Payload)
}

func TestInMemory_Query(t *testing.T) {
	s := newTestMemStore()
	const wildcard = uint32(1815237614)
	zero := time.Unix(0, 0)
	tests := []struct {
		query    []uint32
		limit    int
		count    int
		gathered []byte
	}{
		{query: []uint32{0, 10, 20, 50}, limit: 10, count: 0},
		{query: []uint32{0, 1, 1, 1}, limit: 10, count: 1},
		{query: []uint32{0, 1, 1, wildcard}, limit: 10, count: 2},
		{query: []uint32{0, 1}, limit: 10, count: 6},
		{query: []uint32{0, 2}, limit: 10, count: 0},
		{query: []uint32{0, 1, 2}, limit: 10, count: 2},
		{query: []uint32{0, 1}, limit: 5, count: 5},
		{query: []uint32{0, 1}, limit: 5, count: 5, gathered: []byte{61, 120, 2, 236, 174, 165, 1, 4, 0, 1, 2, 3, 13, 116, 101, 115, 116, 47, 99, 104, 97, 110, 110, 101, 108, 47, 5, 49, 44, 50, 44, 51, 10, 9, 30, 8, 2, 3, 4, 58, 30, 0, 20, 50, 44, 51, 44, 52, 10}},
	}

	for _, tc := range tests {
		if tc.gathered == nil {
			s.survey = nil
		} else {
			s.survey = survey(func(string, []byte) (message.Awaiter, error) {
				return &mockAwaiter{f: func(_ time.Duration) [][]byte { return [][]byte{tc.gathered} }}, nil
			})
		}

		out, err := s.Query(tc.query, zero, zero, tc.limit)
		assert.NoError(t, err)

		count := 0
		for range out {
			count++
		}

		assert.Equal(t, tc.count, count)
	}
}

func TestInMemory_lookup(t *testing.T) {
	s := newTestMemStore()
	const wildcard = uint32(1815237614)
	zero := time.Unix(0, 0)
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
		matches := s.lookup(newLookupQuery(tc.query, zero, zero, tc.limit))
		assert.Equal(t, tc.count, len(matches))
	}
}

func TestInMemory_OnSurvey(t *testing.T) {
	s := NewInMemory(nil)
	s.Configure(nil)
	s.storeFrame(getNTestMessages(10))
	zero := time.Unix(0, 0)
	tests := []struct {
		name        string
		query       lookupQuery
		expectOk    bool
		expectCount int
	}{
		{name: "dummy"},
		{name: "ssdstore"},
		{
			name:        "ssdstore",
			query:       newLookupQuery(message.Ssid{0, 1}, zero, zero, 1),
			expectOk:    true,
			expectCount: 1,
		},
		{
			name:        "ssdstore",
			query:       newLookupQuery(message.Ssid{0, 1}, zero, zero, 10),
			expectOk:    true,
			expectCount: 2,
		},
	}

	for _, tc := range tests {
		q, _ := binary.Marshal(tc.query)
		resp, ok := s.OnSurvey(tc.name, q)
		assert.Equal(t, tc.expectOk, ok)
		if tc.expectOk && ok {
			msgs, err := message.DecodeFrame(resp)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectCount, len(msgs))
		}
	}

	// Special, wrong payload case
	_, ok := s.OnSurvey("ssdstore", []byte{})
	assert.Equal(t, false, ok)

}
