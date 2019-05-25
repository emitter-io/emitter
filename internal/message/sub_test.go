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

package message

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/emitter-io/emitter/internal/security"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSsidEncode(b *testing.B) {
	ssid := NewSsid(0, []uint32{56498455, 44565213, 46512350, 18204498})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ssid.Encode()
	}
}

func TestSsidPresence(t *testing.T) {
	ssid := NewSsidForPresence(Ssid{1, 2, 3})
	assert.NotNil(t, ssid)
	assert.EqualValues(t, Ssid{0, 3869262148, 1, 2, 3}, ssid)
}

func TestSsidShare(t *testing.T) {
	ssid := NewSsidForShare(Ssid{1, 2, 3})
	assert.NotNil(t, ssid)
	assert.EqualValues(t, Ssid{1, share, 2, 3}, ssid)
}

func TestSsid(t *testing.T) {
	c := security.Channel{
		Key:         []byte("key"),
		Channel:     []byte("a/b/c/"),
		Query:       []uint32{10, 20, 50},
		Options:     []security.ChannelOption{},
		ChannelType: security.ChannelStatic,
	}

	ssid := NewSsid(0, c.Query)
	assert.Equal(t, uint32(0), ssid.Contract())
	assert.Equal(t, uint32(0x2c), ssid.GetHashCode())
}

func TestSsidEncode(t *testing.T) {
	tests := []struct {
		ssid     []uint32
		expected string
	}{
		{
			ssid:     []uint32{10, 20, 50},
			expected: "000000000000000a0000001400000032",
		},
		{
			ssid:     []uint32{10, wildcard, 50},
			expected: "000000000000000a........00000032",
		},
	}

	for _, tc := range tests {
		ssid := NewSsid(0, tc.ssid)
		assert.Equal(t, tc.expected, ssid.Encode())
	}
}

func TestSub_NewCounters(t *testing.T) {
	counters := NewCounters()
	assert.NotNil(t, counters.m)
	assert.Empty(t, counters.m)
}

func TestSub_getOrCreate(t *testing.T) {
	// Preparation.
	counters := NewCounters()
	ssid := make([]uint32, 1)
	key := (Ssid(ssid)).GetHashCode()

	// Call.
	createdCounter := counters.getOrCreate(ssid, []byte("test"))

	// Assertions.
	assert.NotEmpty(t, counters.m)

	counter := counters.m[key]
	assert.NotEmpty(t, counter)
	assert.Equal(t, counter, createdCounter)

	assert.Equal(t, counter.Channel, []byte("test"))
	assert.Equal(t, counter.Counter, 0)
	assert.Equal(t, counter.Ssid, Ssid(ssid))
}

func TestSubscribers(t *testing.T) {
	subs := newSubscribers()
	sub := &testSubscriber{id: "x"}

	{
		added := subs.AddUnique(sub)
		assert.True(t, added)
	}
	{
		added := subs.AddUnique(sub)
		assert.False(t, added)
	}
	{
		removed := subs.Remove(sub)
		assert.True(t, removed)
	}
	{
		removed := subs.Remove(sub)
		assert.False(t, removed)
	}
}

func TestSub_All(t *testing.T) {
	// Preparation.
	counters := NewCounters()
	ssid := make([]uint32, 1)
	createdCounter := counters.getOrCreate(ssid, []byte("test"))

	// Call.
	allCounters := counters.All()

	// Assertions.
	assert.Equal(t, 1, len(allCounters))
	assert.Equal(t, createdCounter, &allCounters[0])
}

// TODO : test concurrency
// TODO : add decrement test
func TestSub_Increment(t *testing.T) {
	// Preparation.
	counters := NewCounters()
	ssid1 := make([]uint32, 1)
	ssid2 := make([]uint32, 1)
	ssid2[0] = 1
	key1 := (Ssid(ssid1)).GetHashCode()
	key2 := (Ssid(ssid2)).GetHashCode()

	counters.getOrCreate(ssid1, []byte("test"))

	// Test previously created counter.
	isFirst := counters.Increment(ssid1, []byte("test"))
	assert.True(t, isFirst)
	assert.Equal(t, 1, counters.m[key1].Counter)

	// Test not previously create counter.
	isFirst = counters.Increment(ssid2, []byte("test"))
	assert.True(t, isFirst)
	assert.Equal(t, 1, counters.m[key2].Counter)

	// Test increment previously incremented counter.
	isFirst = counters.Increment(ssid2, []byte("test"))
	assert.False(t, isFirst)
	assert.Equal(t, 2, counters.m[key2].Counter)

	// Test decrement previously incremented counter.
	isDecremented := counters.Decrement(ssid2)
	assert.False(t, isDecremented)
	assert.Equal(t, 1, counters.m[key2].Counter)

	// Test decrement previously incremented counter.
	isDecremented = counters.Decrement(ssid2)
	assert.True(t, isDecremented)
}

func TestCollisions(t *testing.T) {
	subs := newSubscribers()
	count := 100000
	for i := 0; i < count; i++ {
		subs.AddUnique(&testSubscriber{fmt.Sprintf("%d", i)})
	}
	assert.Equal(t, count, subs.Size())
}

func TestRandom(t *testing.T) {
	rand.Seed(42)
	for count := 2; count < 20; count++ {
		testRandom(t, count, 100000)
	}
}

func testRandom(t *testing.T, count, iter int) {
	subs := newSubscribers()
	for i := 0; i < count; i++ {
		subs.AddUnique(&testSubscriber{fmt.Sprintf("%d", i)})
	}

	n := 1552127721834
	x := uint32((n >> 32) ^ n)
	out := make(map[string]int)
	for i := 0; i < iter; i++ {
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		r := x
		if s := subs.Random(r); s != nil {
			out[s.ID()]++
		}
	}

	assert.Equal(t, count, len(out))
	avg := float64(iter) / float64(count)
	for k, v := range out {
		p := (float64(v) - float64(avg)) / float64(avg)
		if p < 0 {
			p = -p
		}
		if p > 0.05 {
			t.Fatalf("[count = %d] skew more than 5%% for key '%v' it's %.2f%%", count, k, p*100)
		}
	}
}

// BenchmarkReset-8   	    2000	    560516 ns/op	     409 B/op	       0 allocs/op
func BenchmarkReset(b *testing.B) {
	orig := newSubscribers()
	subs := newSubscribers()
	for i := 0; i < 10000; i++ {
		orig.AddUnique(&testSubscriber{fmt.Sprintf("%d", i)})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subs.AddRange(orig, nil)
		subs.Reset()
	}
}
