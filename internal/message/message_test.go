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
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestMessage(ssid Ssid, channel, payload string) Message {
	return Message{
		ID:      NewID(ssid),
		Channel: []byte(channel),
		Payload: []byte(payload),
	}
}

func TestDecodeFrame(t *testing.T) {
	frame := Frame{
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/", "hello ab"),
	}

	// Encode
	buffer := frame.Encode()
	assert.True(t, len(buffer) >= 65)

	// Decode
	output, err := DecodeFrame(buffer)
	assert.NoError(t, err)
	assert.Equal(t, frame, output)
}

func TestNewMessage(t *testing.T) {
	m := New(Ssid{1, 2, 3}, []byte("a/b/c/"), []byte("hello abc"))
	assert.Equal(t, int64(9), m.Size())
	assert.Equal(t, Ssid{1, 2, 3}, m.Ssid())
	assert.Equal(t, uint32(1), m.Contract())
	assert.NotNil(t, m.Expires().String())
	assert.False(t, m.Stored())
}

func TestNewFrame(t *testing.T) {
	f := NewFrame(64)
	assert.Len(t, f, 0)
	assert.Equal(t, 64, cap(f))
}

// BenchmarkCodec/Encode-8         	 3788479	       324.9 ns/op	     176 B/op	       1 allocs/op
// BenchmarkCodec/Decode-8         	 3950424	       294.8 ns/op	     288 B/op	       3 allocs/op
func BenchmarkCodec(b *testing.B) {
	m := newTestMessage(Ssid{1, 2, 3}, "tweet/canada/english/", "This is a random tweet en english so we can test the payload. #emitter")
	enc := m.Encode()
	b.Run("Encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Encode()
		}
	})

	b.Run("Decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			DecodeMessage(enc)
		}
	})
}

// BenchmarkEncodeWithSnappy-8   	   10000	    188831 ns/op	   57374 B/op	       1 allocs/op
func BenchmarkEncodeWithSnappy(b *testing.B) {
	var frame Frame
	for m := 0; m < 1000; m++ {
		frame = append(frame, newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame.Encode()
	}
}

// Benchmark_DecodeFrame-8   	    5000	    284238 ns/op	  211217 B/op	    1004 allocs/op
func Benchmark_DecodeFrame(b *testing.B) {
	var frame Frame
	for m := 0; m < 1000; m++ {
		frame = append(frame, newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"))
	}
	encoded := frame.Encode()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeFrame(encoded)
	}
}

func TestFrameLimit(t *testing.T) {
	f := Frame{
		newTestMessage(Ssid{1, 2, 1}, "a/b/a/", "hello aba"),
		newTestMessage(Ssid{1, 2, 2}, "a/b/b/", "hello abb"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 4}, "a/b/d/", "hello abd"),
	}

	f.Limit(2)
	assert.Len(t, f, 2)
	assert.Equal(t, "a/b/c/", string(f[0].Channel))
	assert.Equal(t, "a/b/d/", string(f[1].Channel))
}

func TestFrameSplit(t *testing.T) {
	f := Frame{
		newTestMessage(Ssid{1, 2, 1}, "a/b/a/", "hello aba"),
		newTestMessage(Ssid{1, 2, 2}, "a/b/b/", "hello abb"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 4}, "a/b/d/", "hello abd"),
	}

	head, tail := f.Split(127)
	assert.Len(t, head, 2)
	assert.Len(t, tail, 2)
}

func TestFrameSplit_Empty(t *testing.T) {
	f := Frame{}

	head, tail := f.Split(127)
	assert.Len(t, head, 0)
	assert.Len(t, tail, 0)
}
