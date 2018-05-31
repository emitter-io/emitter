/**********************************************************************************
* Copyright (c) 2009-2018 Misakai Ltd.
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

	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
)

func newTestMessage(ssid Ssid, channel, payload string) Message {
	return Message{
		ID:      NewID(ssid, 0),
		Channel: []byte(channel),
		Payload: []byte(payload),
	}
}

func TestDecodeFrame(t *testing.T) {
	frame := Frame{
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/", "hello ab"),
	}

	// Append
	//frame.Append(0, Ssid{1, 2, 3}, []byte("a/b/c/"), []byte("hello abc"))

	// Encode
	buffer := frame.Encode()
	assert.True(t, len(buffer) >= 65)

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

// 2000000	       577 ns/op	     320 B/op	       2 allocs/op
func BenchmarkEncode(b *testing.B) {
	m := Frame{
		newTestMessage(Ssid{1, 2, 3}, "tweet/canada/english/", "This is a random tweet en english so we can test the payload. #emitter"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Marshal(&m)
	}
}

// 2000000	       799 ns/op	     480 B/op	       3 allocs/op
func BenchmarkEncodeWithSnappy(b *testing.B) {
	m := Frame{
		newTestMessage(Ssid{1, 2, 3}, "tweet/canada/english/", "This is a random tweet en english so we can test the payload. #emitter"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Encode()
	}
}

func TestFrameSort(t *testing.T) {
	f := Frame{
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
	}

	f.Sort()
	/*assert.Equal(t, int64(1), f[0].Time)
	assert.Equal(t, int64(2), f[1].Time)
	assert.Equal(t, int64(3), f[2].Time)
	assert.Equal(t, int64(4), f[3].Time)*/
}

func TestFrameLimit(t *testing.T) {
	f := Frame{
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
	}

	f.Limit(2)
	assert.Len(t, f, 2)
}
