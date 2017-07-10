package broker

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/emitter-io/emitter/utils"
	"github.com/stretchr/testify/assert"
)

func TestTrieMatch(t *testing.T) {
	m := NewSubscriptionTrie()
	testPopulateWithStrings(m, []string{
		"a/",
		"a/b/c/",
		"a/+/c/",
		"a/b/c/d/",
		"a/+/c/+/",
		"x/",
		"x/y/",
		"x/+/z",
	})

	// Tests to run
	tests := []struct {
		topic string
		n     int
	}{
		{topic: "a/", n: 1},
		{topic: "a/1/", n: 1},
		{topic: "a/2/", n: 1},
		{topic: "a/1/2/", n: 1},
		{topic: "a/1/2/3/", n: 1},
		{topic: "a/x/y/c/", n: 1},
		{topic: "a/x/c/", n: 2},
		{topic: "a/b/c/", n: 3},
		{topic: "a/b/c/d/", n: 5},
		{topic: "a/b/c/e/", n: 4},
		{topic: "x/y/c/e/", n: 2},
	}

	for _, tc := range tests {
		result := m.Lookup(testSub(tc.topic))
		assert.Equal(t, tc.n, len(result))
	}
}

func TestTrieIntegration(t *testing.T) {
	assert := assert.New(t)
	var (
		m  = NewSubscriptionTrie()
		s0 = new(Conn)
		s1 = new(Conn)
		s2 = new(Conn)
	)

	sub0, err := m.Subscribe([]uint32{1, wildcard}, "", s0)
	assert.NoError(err)
	sub1, err := m.Subscribe([]uint32{wildcard, 2}, "", s0)
	assert.NoError(err)
	sub2, err := m.Subscribe([]uint32{1, 3}, "", s0)
	assert.NoError(err)
	sub3, err := m.Subscribe([]uint32{wildcard, 3}, "", s1)
	assert.NoError(err)
	sub4, err := m.Subscribe([]uint32{1, wildcard}, "", s1)
	assert.NoError(err)
	sub5, err := m.Subscribe([]uint32{4}, "", s1)
	assert.NoError(err)
	sub6, err := m.Subscribe([]uint32{wildcard}, "", s2)
	assert.NoError(err)

	assertEqual(assert, Subscribers{s0, s1, s2}, m.Lookup([]uint32{1, 3}))
	assertEqual(assert, Subscribers{s2}, m.Lookup([]uint32{1}))
	assertEqual(assert, Subscribers{s1, s2}, m.Lookup([]uint32{4, 5}))
	assertEqual(assert, Subscribers{s0, s1, s2}, m.Lookup([]uint32{1, 5}))
	assertEqual(assert, Subscribers{s1, s2}, m.Lookup([]uint32{4}))

	m.Unsubscribe(sub0)
	m.Unsubscribe(sub1)
	m.Unsubscribe(sub2)
	m.Unsubscribe(sub3)
	m.Unsubscribe(sub4)
	m.Unsubscribe(sub5)
	m.Unsubscribe(sub6)

	assertEqual(assert, []Subscriber{}, m.Lookup([]uint32{1, 3}))
	assertEqual(assert, []Subscriber{}, m.Lookup([]uint32{1}))
	assertEqual(assert, []Subscriber{}, m.Lookup([]uint32{4, 5}))
	assertEqual(assert, []Subscriber{}, m.Lookup([]uint32{1, 5}))
	assertEqual(assert, []Subscriber{}, m.Lookup([]uint32{4}))
}

// Populates the trie with a set of strings
func testPopulateWithStrings(m *SubscriptionTrie, values []string) {
	for _, s := range values {
		m.Subscribe(testSub(s), s, new(Conn))
	}
}

// TestSub gets the SSID from a string, making it a bit easier to read.
func testSub(topic string) []uint32 {
	ssid := make([]uint32, 0)
	for _, p := range strings.Split(topic, "/") {
		if len(p) > 0 {
			ssid = append(ssid, utils.GetHash([]byte(p)))
		}
	}
	return ssid
}

func BenchmarkSubscriptionTrieSubscribe(b *testing.B) {
	var (
		m     = NewSubscriptionTrie()
		s0    = new(Conn)
		query = []uint32{1, wildcard, 2, 3, 4}
	)
	populateMatcher(m, 1000, 5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Subscribe(query, "", s0)
	}
}

func BenchmarkSubscriptionTrieUnsubscribe(b *testing.B) {
	var (
		m     = NewSubscriptionTrie()
		s0    = new(Conn)
		query = []uint32{1, wildcard, 2, 3, 4}
	)

	id, _ := m.Subscribe(query, "", s0)
	populateMatcher(m, 1000, 5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Unsubscribe(id)
	}
}

func BenchmarkSubscriptionTrieLookup(b *testing.B) {
	var (
		m  = NewSubscriptionTrie()
		s0 = new(Conn)
		q1 = []uint32{1, wildcard, 2, 3, 4}
		q2 = []uint32{1, 5, 2, 3, 4}
	)

	m.Subscribe(q1, "", s0)
	populateMatcher(m, 1000, 5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lookup(q2)
	}
}

func BenchmarkSubscriptionTrieSubscribeCold(b *testing.B) {
	var (
		m     = NewSubscriptionTrie()
		s0    = new(Conn)
		query = []uint32{1, wildcard, 2, 3, 4}
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Subscribe(query, "", s0)
	}
}

func BenchmarkSubscriptionTrieUnsubscribeCold(b *testing.B) {
	var (
		m     = NewSubscriptionTrie()
		s0    = new(Conn)
		query = []uint32{1, wildcard, 2, 3, 4}
	)
	id, _ := m.Subscribe(query, "", s0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Unsubscribe(id)
	}
}

func BenchmarkSubscriptionTrieLookupCold(b *testing.B) {
	var (
		m  = NewSubscriptionTrie()
		s0 = new(Conn)
		q1 = []uint32{1, wildcard, 2, 3, 4}
		q2 = []uint32{1, 5, 2, 3, 4}
	)
	m.Subscribe(q1, "", s0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lookup(q2)
	}
}

func assertEqual(assert *assert.Assertions, expected, actual Subscribers) {
	assert.Len(actual, len(expected))
	for _, sub := range actual {
		assert.True(expected.Contains(sub))
	}
}

func populateMatcher(m *SubscriptionTrie, num, topicSize int) {
	for i := 0; i < num; i++ {
		topic := make([]uint32, 0)
		for j := 0; j < topicSize; j++ {
			topic = append(topic, uint32(rand.Int()))
		}

		m.Subscribe(topic, "", new(Conn))
	}
}
