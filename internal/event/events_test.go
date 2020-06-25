/**********************************************************************************
* Copyright (c) 2009-2020 Misakai Ltd.
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

package event

import (
	"testing"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security/hash"
	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
)

func TestEncodeSubscription(t *testing.T) {
	ev := Subscription{
		Ssid:    message.Ssid{hash.OfString("a"), hash.OfString("b"), hash.OfString("c"), hash.OfString("d"), hash.OfString("e")},
		Peer:    657,
		Conn:    12456,
		User:    "hello",
		Channel: []byte("a/b/c/d/e/"),
	}

	assert.Equal(t, "LPCGOQV6DEDQFHIRWBMKICQCZE", ev.ConnID())

	// Encode
	k, v := ev.Key(), ev.Val()
	assert.Equal(t, typeSub, ev.unitType())
	assert.Equal(t, 36, len(k))
	assert.Equal(t,
		[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x91, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30, 0xa8, 0xc1, 0x3, 0xea,
			0xb3, 0x1d, 0xd8, 0x2e, 0x48, 0x3d, 0x43, 0x19, 0x23, 0x16, 0x1b, 0xa0, 0x5b, 0x7f, 0xd4, 0x1, 0x2},
		[]byte(k),
	)

	// Decode
	dec, err := decodeSubscription(k, v)
	assert.NoError(t, err)
	assert.Equal(t, ev, dec)
}

func TestEncodeBan(t *testing.T) {
	ev := Ban("a/b/c/d/e/")
	assert.Nil(t, ev.Val())

	// Encode
	enc := ev.Key()
	assert.Equal(t, typeBan, ev.unitType())
	assert.Equal(t, 10, len(enc))
	assert.Equal(t,
		[]byte{0x61, 0x2f, 0x62, 0x2f, 0x63, 0x2f, 0x64, 0x2f, 0x65, 0x2f},
		[]byte(enc),
	)

	// Decode
	dec, err := decodeBan(enc)
	assert.NoError(t, err)
	assert.Equal(t, ev, dec)
}

func TestEncodeConnection(t *testing.T) {
	ev := Connection{
		Peer:        657,
		Conn:        12456,
		WillFlag:    true,
		WillRetain:  true,
		WillQoS:     1,
		WillTopic:   binary.ToBytes("a/b/c/d/"),
		WillMessage: binary.ToBytes("hello"),
		ClientID:    binary.ToBytes("client id"),
		Username:    binary.ToBytes("username"),
	}

	// Encode
	k, v := ev.Key(), ev.Val()
	assert.Equal(t, typeConn, ev.unitType())
	assert.Equal(t, 16, len(k))
	assert.Equal(t,
		[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x91, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30, 0xa8},
		[]byte(k),
	)
	assert.Equal(t, []byte{
		0x1, 0x1, 0x1, 0x8, 0x61, 0x2f, 0x62, 0x2f, 0x63, 0x2f, 0x64, 0x2f, 0x5, 0x68,
		0x65, 0x6c, 0x6c, 0x6f, 0x9, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x20, 0x69,
		0x64, 0x8, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65},
		v)

	// Decode
	dec, err := decodeConnection(k, v)
	assert.NoError(t, err)
	assert.Equal(t, ev, dec)
}

// Benchmark_Subscription/encode-8         	 5939726	       199 ns/op	     160 B/op	       3 allocs/op
// Benchmark_Subscription/decode-8         	 6665554	       178 ns/op	     112 B/op	       2 allocs/op
func Benchmark_Subscription(b *testing.B) {
	ev := Subscription{
		Ssid:    message.Ssid{1, 2, 3, 4, 5},
		Peer:    657,
		Conn:    12456,
		User:    "hello",
		Channel: []byte("a/b/c/d/e/"),
	}

	// Encode
	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ev.Key()
			ev.Val()
		}
	})

	// Decode
	k, v := ev.Key(), ev.Val()
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			decodeSubscription(k, v)
		}
	})

}
