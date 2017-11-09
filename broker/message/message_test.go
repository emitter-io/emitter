package message

import (
	"testing"

	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/emitter-io/emitter/utils"
	"github.com/stretchr/testify/assert"
)

func TestDecodeFrame(t *testing.T) {
	frame := Frame{
		Message{Ssid: subscription.Ssid{1, 2, 3}, Channel: []byte("a/b/c/"), Payload: []byte("hello abc")},
		Message{Ssid: subscription.Ssid{1, 2, 3}, Channel: []byte("a/b/"), Payload: []byte("hello ab")},
	}

	// Encode
	buffer, err := frame.Encode()
	assert.NoError(t, err)

	// Decode
	output, err := DecodeFrame(buffer)
	assert.NoError(t, err)
	assert.Equal(t, frame, output)
}

func TestFrameAppend(t *testing.T) {
	var frame Frame
	frame.Append(1111, subscription.Ssid{1, 2, 3}, []byte("a/b/c/"), []byte("hello abc"))
	frame.Append(1111, subscription.Ssid{1, 2, 3}, []byte("a/b/c/"), []byte("hello abc"))
	assert.Len(t, frame, 2)
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
