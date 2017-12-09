package message

import (
	"testing"

	"github.com/emitter-io/emitter/utils"
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
	buffer, err := frame.Encode()
	assert.NoError(t, err)

	// Decode
	output, err := DecodeFrame(buffer)
	assert.NoError(t, err)
	assert.Equal(t, frame, output)
}

func TestMessageSize(t *testing.T) {
	m := Message{Payload: []byte("hello abc")}
	assert.Equal(t, int64(9), m.Size())
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
		utils.Encode(&m)
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
