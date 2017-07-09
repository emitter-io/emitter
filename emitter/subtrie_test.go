package emitter

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
		s0 = 0
		s1 = 1
		s2 = 2
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
		m.Subscribe(testSub(s), s, s)
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

/*
func TestTrieRaces(t *testing.T) {
	trie := NewSubscriptionTrie()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			id := Ssid([]uint32{1, 2, 3})
			s := &Subscription{Ssid: id}

			// We should be able to add it
			assert.True(t, true, trie.Add(s))

			// We should be able to retrieve it now
			o, _ := trie.Get(id)
			assert.Equal(t, s, o)

			// We should be able to remove it
			assert.True(t, true, trie.Remove(s))
			wg.Done()
		}()
	}

	// Wait
	wg.Wait()
}
*/

func BenchmarkSubscriptionTrieSubscribe(b *testing.B) {
	var (
		m     = NewSubscriptionTrie()
		s0    = 0
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
		s0    = 0
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
		s0 = 0
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
		s0    = 0
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
		s0    = 0
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
		s0 = 0
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

func assertEqual(assert *assert.Assertions, expected, actual []Subscriber) {
	assert.Len(actual, len(expected))
	for _, sub := range expected {
		assert.Contains(actual, sub)
	}
}

func populateMatcher(m *SubscriptionTrie, num, topicSize int) {
	for i := 0; i < num; i++ {
		topic := make([]uint32, 0)
		for j := 0; j < topicSize; j++ {
			topic = append(topic, uint32(rand.Int()))
		}

		m.Subscribe(topic, "", Subscriber(1))
	}
}

/*func benchmark5050(b *testing.B, numItems, numThreads int, factory func([][]string) Matcher) {
	itemsToInsert := make([][]string, 0, numThreads)
	for i := 0; i < numThreads; i++ {
		items := make([]string, 0, numItems)
		for j := 0; j < numItems; j++ {
			topic := strconv.Itoa(j%10) + "." + strconv.Itoa(j%50) + "." + strconv.Itoa(j)
			items = append(items, topic)
		}
		itemsToInsert = append(itemsToInsert, items)
	}

	var wg sync.WaitGroup
	sub := Subscriber("abc")
	m := factory(itemsToInsert)
	populateMatcher(m, 1000, 5)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(numThreads)
		for j := 0; j < numThreads; j++ {
			go func(j int) {
				if j%2 != 0 {
					for _, key := range itemsToInsert[j] {
						m.Subscribe(key, sub)
					}
				} else {
					for _, key := range itemsToInsert[j] {
						m.Lookup(key)
					}
				}
				wg.Done()
			}(j)
		}
		wg.Wait()
	}
}

func benchmark9010(b *testing.B, numItems, numThreads int, factory func([][]string) Matcher) {
	itemsToInsert := make([][]string, 0, numThreads)
	for i := 0; i < numThreads; i++ {
		items := make([]string, 0, numItems)
		for j := 0; j < numItems; j++ {
			topic := strconv.Itoa(j%10) + "." + strconv.Itoa(j%50) + "." + strconv.Itoa(j)
			items = append(items, topic)
		}
		itemsToInsert = append(itemsToInsert, items)
	}

	var wg sync.WaitGroup
	sub := Subscriber("abc")
	m := factory(itemsToInsert)
	populateMatcher(m, 1000, 5)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(numThreads)
		for j := 0; j < numThreads; j++ {
			go func(j int) {
				if j%10 == 0 {
					for _, key := range itemsToInsert[j] {
						m.Subscribe(key, sub)
					}
				} else {
					for _, key := range itemsToInsert[j] {
						m.Lookup(key)
					}
				}
				wg.Done()
			}(j)
		}
		wg.Wait()
	}
}
*/

/*
func BenchmarkMultithreaded1Thread5050CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 1
	benchmark5050(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded2Thread5050CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 2
	benchmark5050(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded4Thread5050CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 4
	benchmark5050(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded8Thread5050CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 8
	benchmark5050(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded12Thread5050CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 12
	benchmark5050(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded16Thread5050CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 16
	benchmark5050(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded1Thread9010CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 1
	benchmark9010(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded2Thread9010CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 2
	benchmark9010(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded4Thread9010CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 4
	benchmark9010(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded8Thread9010CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 8
	benchmark9010(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded12Thread9010CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 12
	benchmark9010(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}

func BenchmarkMultithreaded16Thread9010CSTrie(b *testing.B) {
	numItems := 1000
	numThreads := 16
	benchmark9010(b, numItems, numThreads, func(items [][]string) Matcher {
		return NewCSTrieMatcher()
	})
}
*/
