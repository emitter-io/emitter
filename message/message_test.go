package message

import (
	"testing"

	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
)

func TestDecodeFrame(t *testing.T) {
	frame := Frame{
		Message{Ssid: Ssid{1, 2, 3}, Channel: []byte("a/b/c/"), Payload: []byte("hello abc")},
		Message{Ssid: Ssid{1, 2, 3}, Channel: []byte("a/b/"), Payload: []byte("hello ab")},
	}

	// Append
	frame.Append(0, Ssid{1, 2, 3}, []byte("a/b/c/"), []byte("hello abc"))

	// Encode
	buffer := frame.Encode()
	assert.Len(t, buffer, 41)

	// Decode
	output, err := DecodeFrame(buffer)
	assert.NoError(t, err)
	assert.Equal(t, frame, output)
}

func TestMessageSize(t *testing.T) {
	m := Message{Payload: []byte("hello abc")}
	assert.Equal(t, int64(9), m.Size())
}

func TestNewFrame(t *testing.T) {
	f := NewFrame(64)
	assert.Len(t, f, 0)
	assert.Equal(t, 64, cap(f))
}

func BenchmarkEncode(b *testing.B) {
	m := Frame{{
		Time:    1234,
		Channel: []byte("tweet/canada/english/"),
		Payload: []byte("This is a random tweet en english so we can test the payload. #emitter"),
		Ssid:    []uint32{1, 2, 3},
	}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Marshal(&m)
	}
}

func BenchmarkEncodeWithSnappy(b *testing.B) {
	m := Frame{{
		Time:    1234,
		Channel: []byte("tweet/canada/english/"),
		Payload: []byte("This is a random tweet en english so we can test the payload. #emitter"),
		Ssid:    []uint32{1, 2, 3},
	}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Encode()
	}
}

func TestFrameSort(t *testing.T) {
	f := Frame{
		Message{Time: 2},
		Message{Time: 1},
		Message{Time: 4},
		Message{Time: 3},
	}

	f.Sort()
	assert.Equal(t, int64(1), f[0].Time)
	assert.Equal(t, int64(2), f[1].Time)
	assert.Equal(t, int64(3), f[2].Time)
	assert.Equal(t, int64(4), f[3].Time)
}

func TestFrameLimit(t *testing.T) {
	f := Frame{
		Message{Time: 2},
		Message{Time: 4},
		Message{Time: 1},
		Message{Time: 3},
	}

	f.Limit(2)
	assert.Equal(t, int64(2), f[0].Time)
	assert.Equal(t, int64(4), f[1].Time)
	assert.Len(t, f, 2)
}
