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
	"strings"
	"testing"

	"github.com/emitter-io/emitter/internal/security/hash"
	"github.com/stretchr/testify/assert"
)

type testSubscriber struct {
	id string
}

func (s *testSubscriber) ID() string {
	return s.id
}

func (s *testSubscriber) Type() SubscriberType {
	return SubscriberDirect
}

func (s *testSubscriber) Send(*Message) error {
	return nil
}

func TestTrieMatch1(t *testing.T) {
	m := NewTrie()
	testPopulateWithStrings(m, []string{
		"a/",
	})

	// Tests to run
	tests := []struct {
		topic string
		n     int
	}{
		{topic: "a/b/", n: 1},
	}

	for _, tc := range tests {
		result := m.Lookup(testSub(tc.topic), nil)
		assert.Equal(t, tc.n, result.Size())
	}
}

func TestTrieMatch(t *testing.T) {
	m := NewTrie()
	testPopulateWithStrings(m, []string{
		"key/a/",
		"key/a/b/c/",
		"key/a/+/c/",
		"key/a/b/c/d/",
		"key/a/+/c/+/",
		"key/x/",
		"key/x/y/",
		"key/x/+/z",
		"key/$share/group1/a/+/c/",
		"key/$share/group1/a/b/c/",
		"key/$share/group2/a/b/c/",
		"key/$share/group2/a/b/",
		"key/$share/group3/y/",
		"key/$share/group3/y/", // Duplicate (same ID)
	})

	// Tests to run
	tests := []struct {
		topic string
		n     int
	}{
		{topic: "key/a/", n: 1},
		{topic: "key/a/1/", n: 1},
		{topic: "key/a/2/", n: 1},
		{topic: "key/a/1/2/", n: 1},
		{topic: "key/a/1/2/3/", n: 1},
		{topic: "key/a/x/y/c/", n: 1},
		{topic: "key/a/x/c/", n: 3},
		{topic: "key/a/b/c/", n: 5},
		{topic: "key/a/b/c/d/", n: 7},
		{topic: "key/a/b/c/e/", n: 6},
		{topic: "key/x/y/c/e/", n: 2},
		{topic: "key/y/", n: 1},
	}

	assert.Equal(t, 13, m.Count())
	for _, tc := range tests {
		result := m.Lookup(testSub(tc.topic), nil)
		assert.Equal(t, tc.n, result.Size(), tc.topic)
	}
}

func TestTrieMatchMqtt0(t *testing.T) {
	m := NewTrieMQTT()
	testPopulateWithStrings(m, []string{
		"a/",
	})

	// Tests to run
	tests := []struct {
		topic string
		n     int
	}{
		{topic: "a/b/", n: 0},
	}

	for _, tc := range tests {
		result := m.Lookup(testSub(tc.topic), nil)
		assert.Equal(t, tc.n, result.Size())
	}
}

func TestTrieMatchMqttMultiWildcard(t *testing.T) {
	m := NewTrieMQTT()
	testPopulateWithStrings(m, []string{
		"a/#",
		"+/b/#",
	})

	// Tests to run
	tests := []struct {
		topic string
		n     int
	}{
		{topic: "a/", n: 0},
		{topic: "a/b/", n: 1},
		{topic: "a/b/c/", n: 2},
		{topic: "a/b/c/d/", n: 2},
	}

	for _, tc := range tests {
		result := m.Lookup(testSub(tc.topic), nil)
		assert.Equal(t, tc.n, result.Size())
	}
}

func TestTrieMatchMqtt(t *testing.T) {
	m := NewTrieMQTT()
	testPopulateWithStrings(m, []string{
		"key/a/",
		"key/a/b/c/",
		"key/a/+/c/",
		"key/a/b/c/d/",
		"key/a/+/c/+/",
		"key/x/",
		"key/x/y/",
		"key/x/+/z",
		"key/$share/group1/a/+/c/",
		"key/$share/group1/a/b/c/",
		"key/$share/group2/a/b/c/",
		"key/$share/group2/a/b/",
		"key/$share/group3/y/",
		"key/$share/group3/y/", // Duplicate (same ID)
	})

	// Tests to run
	tests := []struct {
		topic string
		n     int
	}{
		{topic: "key/a/", n: 1},
		{topic: "key/a/1/", n: 0},
		{topic: "key/a/2/", n: 0},
		{topic: "key/a/1/2/", n: 0},
		{topic: "key/a/1/2/3/", n: 0},
		{topic: "key/a/x/y/c/", n: 0},
		{topic: "key/a/x/c/", n: 2},
		{topic: "key/a/b/c/", n: 4},
		{topic: "key/a/b/c/d/", n: 2},
		{topic: "key/a/b/c/e/", n: 1},
		{topic: "key/x/y/c/e/", n: 0},
		{topic: "key/y/", n: 1},
	}

	assert.Equal(t, 13, m.Count())
	for _, tc := range tests {
		result := m.Lookup(testSub(tc.topic), nil)
		assert.Equal(t, tc.n, result.Size(), tc.topic)
	}
}

func TestTrieIntegration(t *testing.T) {
	assert := assert.New(t)
	var (
		m  = NewTrie()
		s0 = &testSubscriber{"s0"}
		s1 = &testSubscriber{"s1"}
		s2 = &testSubscriber{"s2"}
	)

	sub0 := m.Subscribe([]uint32{1, wildcard}, s0)
	sub1 := m.Subscribe([]uint32{wildcard, 2}, s0)
	sub2 := m.Subscribe([]uint32{1, 3}, s0)
	sub3 := m.Subscribe([]uint32{wildcard, 3}, s1)
	sub4 := m.Subscribe([]uint32{1, wildcard}, s1)
	sub5 := m.Subscribe([]uint32{4}, s1)
	sub6 := m.Subscribe([]uint32{wildcard}, s2)
	m.Subscribe([]uint32{wildcard}, s2)

	assertEqual(assert, m.Lookup([]uint32{1, 3}, nil), s0, s1, s2)
	assertEqual(assert, m.Lookup([]uint32{1}, nil), s2)
	assertEqual(assert, m.Lookup([]uint32{4, 5}, nil), s1, s2)
	assertEqual(assert, m.Lookup([]uint32{1, 5}, nil), s0, s1, s2)
	assertEqual(assert, m.Lookup([]uint32{4}, nil), s1, s2)

	m.Unsubscribe(sub0.Ssid, sub0.Subscriber)
	m.Unsubscribe(sub1.Ssid, sub1.Subscriber)
	m.Unsubscribe(sub2.Ssid, sub2.Subscriber)
	m.Unsubscribe(sub3.Ssid, sub3.Subscriber)
	m.Unsubscribe(sub4.Ssid, sub4.Subscriber)
	m.Unsubscribe(sub5.Ssid, sub5.Subscriber)
	m.Unsubscribe(sub6.Ssid, sub6.Subscriber)
	m.Unsubscribe(sub6.Ssid, sub6.Subscriber)

	assertEqual(assert, m.Lookup([]uint32{1, 3}, nil))
	assertEqual(assert, m.Lookup([]uint32{1}, nil))
	assertEqual(assert, m.Lookup([]uint32{4, 5}, nil))
	assertEqual(assert, m.Lookup([]uint32{1, 5}, nil))
	assertEqual(assert, m.Lookup([]uint32{4}, nil))
}

func TestTrieIntegrationMqtt(t *testing.T) {
	assert := assert.New(t)
	var (
		m  = NewTrieMQTT()
		s0 = &testSubscriber{"s0"}
		s1 = &testSubscriber{"s1"}
		s2 = &testSubscriber{"s2"}
	)

	sub0 := m.Subscribe([]uint32{1, wildcard}, s0)
	sub1 := m.Subscribe([]uint32{wildcard, 2}, s0)
	sub2 := m.Subscribe([]uint32{1, 3}, s0)
	sub3 := m.Subscribe([]uint32{wildcard, 3}, s1)
	sub4 := m.Subscribe([]uint32{1, wildcard}, s1)
	sub5 := m.Subscribe([]uint32{4}, s1)
	sub6 := m.Subscribe([]uint32{wildcard}, s2)
	m.Subscribe([]uint32{wildcard}, s2)

	assertEqual(assert, m.Lookup([]uint32{1, 3}, nil), s0, s1)
	assertEqual(assert, m.Lookup([]uint32{1}, nil), s2)
	assertEqual(assert, m.Lookup([]uint32{4, 5}, nil))
	assertEqual(assert, m.Lookup([]uint32{1, 5}, nil), s0, s1)
	assertEqual(assert, m.Lookup([]uint32{4}, nil), s1, s2)

	m.Unsubscribe(sub0.Ssid, sub0.Subscriber)
	m.Unsubscribe(sub1.Ssid, sub1.Subscriber)
	m.Unsubscribe(sub2.Ssid, sub2.Subscriber)
	m.Unsubscribe(sub3.Ssid, sub3.Subscriber)
	m.Unsubscribe(sub4.Ssid, sub4.Subscriber)
	m.Unsubscribe(sub5.Ssid, sub5.Subscriber)
	m.Unsubscribe(sub6.Ssid, sub6.Subscriber)
	m.Unsubscribe(sub6.Ssid, sub6.Subscriber)

	assertEqual(assert, m.Lookup([]uint32{1, 3}, nil))
	assertEqual(assert, m.Lookup([]uint32{1}, nil))
	assertEqual(assert, m.Lookup([]uint32{4, 5}, nil))
	assertEqual(assert, m.Lookup([]uint32{1, 5}, nil))
	assertEqual(assert, m.Lookup([]uint32{4}, nil))
}

// Populates the trie with a set of strings
func testPopulateWithStrings(m *Trie, values []string) {
	for _, s := range values {
		m.Subscribe(testSub(s), &testSubscriber{s})
	}
}

// TestSub gets the SSID from a string, making it a bit easier to read.
func testSub(topic string) []uint32 {
	ssid := make([]uint32, 0)
	for _, p := range strings.Split(topic, "/") {
		if len(p) > 0 {
			ssid = append(ssid, hash.Of([]byte(p)))
		}
	}
	return ssid
}

func BenchmarkSubscriptionTrieSubscribe(b *testing.B) {
	var (
		m     = NewTrie()
		s0    = new(testSubscriber)
		query = []uint32{1, wildcard, 2, 3, 4}
	)
	populateMatcher(m, 1000, 5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Subscribe(query, s0)
	}
}

func BenchmarkSubscriptionTrieUnsubscribe(b *testing.B) {
	var (
		m     = NewTrie()
		s0    = new(testSubscriber)
		query = []uint32{1, wildcard, 2, 3, 4}
	)

	m.Subscribe(query, s0)
	populateMatcher(m, 1000, 5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Unsubscribe(query, s0)
	}
}

// BenchmarkTrieLargeN-8   	      50	  24884968 ns/op	    5163 B/op	       8 allocs/op
// BenchmarkTrieLargeN-8   	    2000	    637117 ns/op	   12309 B/op	      10 allocs/op
// BenchmarkTrieLargeN-8   	    3000	    542881 ns/op	   12277 B/op	      10 allocs/op
func BenchmarkTrieLargeN(b *testing.B) {
	rand.Seed(42)
	var (
		m     = NewTrie()
		query = []uint32{1, 2, 3}
	)

	populateMatcher(m, 100000, 3)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lookup(query, nil)
	}
}

// BenchmarkSubscriptionTrieLookup-8   	  200000	     11055 ns/op	    5072 B/op	      52 allocs/op
// BenchmarkSubscriptionTrieLookup-8   	  200000	      7106 ns/op	    1504 B/op	      11 allocs/op
// BenchmarkSubscriptionTrieLookup-8   	  200000	      5415 ns/op	     528 B/op	       6 allocs/op
// BenchmarkSubscriptionTrieLookup-8   	  300000	      4940 ns/op	     496 B/op	       5 allocs/op
// BenchmarkSubscriptionTrieLookup-8   	  200000	      9525 ns/op	     755 B/op	       2 allocs/op
// BenchmarkSubscriptionTrieLookup-8   	  200000	      8183 ns/op	     752 B/op	       2 allocs/op
func BenchmarkSubscriptionTrieLookup(b *testing.B) {
	rand.Seed(42)
	var (
		m  = NewTrie()
		s0 = new(testSubscriber)
		q1 = []uint32{1, wildcard, 2, 3, 4}
		q2 = []uint32{1, 5, 2, 3, 4}
	)

	m.Subscribe(q1, s0)
	populateMatcher(m, 1000, 3)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lookup(q2, nil)
	}
}

func BenchmarkSubscriptionTrieSubscribeCold(b *testing.B) {
	var (
		m     = NewTrie()
		s0    = new(testSubscriber)
		query = []uint32{1, wildcard, 2, 3, 4}
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Subscribe(query, s0)
	}
}

func BenchmarkSubscriptionTrieUnsubscribeCold(b *testing.B) {
	var (
		m     = NewTrie()
		s0    = new(testSubscriber)
		query = []uint32{1, wildcard, 2, 3, 4}
	)
	m.Subscribe(query, s0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Unsubscribe(query, s0)
	}
}

func BenchmarkSubscriptionTrieLookupCold(b *testing.B) {
	var (
		m  = NewTrie()
		s0 = new(testSubscriber)
		q1 = []uint32{1, wildcard, 2, 3, 4}
		q2 = []uint32{1, 5, 2, 3, 4}
	)
	m.Subscribe(q1, s0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lookup(q2, nil)
	}
}
func BenchmarkSubscriptionTrieSubscribeMqtt(b *testing.B) {
	var (
		m     = NewTrieMQTT()
		s0    = new(testSubscriber)
		query = []uint32{1, wildcard, 2, 3, 4}
	)
	populateMatcher(m, 1000, 5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Subscribe(query, s0)
	}
}

func BenchmarkSubscriptionTrieUnsubscribeMqtt(b *testing.B) {
	var (
		m     = NewTrieMQTT()
		s0    = new(testSubscriber)
		query = []uint32{1, wildcard, 2, 3, 4}
	)

	m.Subscribe(query, s0)
	populateMatcher(m, 1000, 5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Unsubscribe(query, s0)
	}
}

func BenchmarkTrieLargeNMqtt(b *testing.B) {
	rand.Seed(42)
	var (
		m     = NewTrieMQTT()
		query = []uint32{1, 2, 3}
	)

	populateMatcher(m, 100000, 3)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lookup(query, nil)
	}
}

func BenchmarkSubscriptionTrieLookupMqtt(b *testing.B) {
	rand.Seed(42)
	var (
		m  = NewTrieMQTT()
		s0 = new(testSubscriber)
		q1 = []uint32{1, wildcard, 2, 3, 4}
		q2 = []uint32{1, 5, 2, 3, 4}
	)

	m.Subscribe(q1, s0)
	populateMatcher(m, 1000, 3)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lookup(q2, nil)
	}
}

func BenchmarkSubscriptionTrieSubscribeColdMqtt(b *testing.B) {
	var (
		m     = NewTrieMQTT()
		s0    = new(testSubscriber)
		query = []uint32{1, wildcard, 2, 3, 4}
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Subscribe(query, s0)
	}
}

func BenchmarkSubscriptionTrieUnsubscribeColdMqtt(b *testing.B) {
	var (
		m     = NewTrieMQTT()
		s0    = new(testSubscriber)
		query = []uint32{1, wildcard, 2, 3, 4}
	)
	m.Subscribe(query, s0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Unsubscribe(query, s0)
	}
}

func BenchmarkSubscriptionTrieLookupColdMqtt(b *testing.B) {
	var (
		m  = NewTrieMQTT()
		s0 = new(testSubscriber)
		q1 = []uint32{1, wildcard, 2, 3, 4}
		q2 = []uint32{1, 5, 2, 3, 4}
	)
	m.Subscribe(q1, s0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lookup(q2, nil)
	}
}

func assertEqual(assert *assert.Assertions, actual Subscribers, expected ...Subscriber) {
	assert.Equal(actual.Size(), len(expected))
	for _, sub := range expected {
		assert.True(actual.Contains(sub))
	}
}

func populateMatcher(m *Trie, num, topicSize int) {
	for i := 0; i < num; i++ {
		topic := make([]uint32, 0)
		for j := 0; j < topicSize; j++ {
			topic = append(topic, uint32(rand.Intn(10)))
		}

		// Add a normal subscriber
		m.Subscribe(topic, &testSubscriber{fmt.Sprintf("%d", rand.Int())})

		// Add a share subscriber
		topic[1] = share
		m.Subscribe(topic, &testSubscriber{fmt.Sprintf("%d", rand.Int())})
	}
}
