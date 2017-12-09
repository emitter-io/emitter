package message

import (
	"testing"

	"github.com/emitter-io/emitter/security"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSsidEncode(b *testing.B) {
	ssid := NewSsid(0, &security.Channel{Query: []uint32{56498455, 44565213, 46512350, 18204498}})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ssid.Encode()
	}
}

func TestSsid(t *testing.T) {
	c := security.Channel{
		Key:         []byte("key"),
		Channel:     []byte("a/b/c/"),
		Query:       []uint32{10, 20, 50},
		Options:     []security.ChannelOption{},
		ChannelType: security.ChannelStatic,
	}

	ssid := NewSsid(0, &c)
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
		ssid := NewSsid(0, &security.Channel{Query: tc.ssid})
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

func TestSub_All(t *testing.T) {
	// Preparation.
	counters := NewCounters()
	ssid := make([]uint32, 1)
	createdCounter := counters.getOrCreate(ssid, []byte("test"))

	// Call.
	allCounters := counters.All()

	// Assertions.
	assert.Equal(t, 1, len(allCounters))

	// TODO : just don't know what I'm doing... http://reactiongifs.me/wp-content/uploads/2013/08/house-pretend-to-work-now.gif
	// tired, will try again another time. (pointers)

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
