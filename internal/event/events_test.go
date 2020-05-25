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
	"github.com/stretchr/testify/assert"
)

func TestEncodeSubscription(t *testing.T) {
	ev := Subscription{
		Ssid:    message.Ssid{1, 2, 3, 4, 5},
		Peer:    657,
		Conn:    12456,
		User:    "hello",
		Channel: []byte("a/b/c/d/e/"),
	}

	assert.Equal(t, "LPCGOQV6DEDQFHIRWBMKICQCZE", ev.ConnID())

	// Encode
	enc := ev.Encode()
	assert.Equal(t, typeSub, ev.unitType())
	assert.Equal(t, 27, len(enc))
	assert.Equal(t,
		[]byte{0x91, 0x5, 0xa8, 0x61, 0x5, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0xa, 0x61, 0x2f, 0x62, 0x2f, 0x63, 0x2f, 0x64, 0x2f, 0x65, 0x2f, 0x5, 0x1, 0x2, 0x3, 0x4, 0x5},
		[]byte(enc),
	)

	// Decode
	dec, err := decodeSubscription(enc)
	assert.NoError(t, err)
	assert.Equal(t, ev, dec)
}

func TestEncodeBan(t *testing.T) {
	ev := Ban("a/b/c/d/e/")

	// Encode
	enc := ev.Encode()
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

// Benchmark_Subscription/encode-8         	 4379755	       270 ns/op	     112 B/op	       2 allocs/op
// Benchmark_Subscription/decode-8         	 2803533	       428 ns/op	     176 B/op	       4 allocs/op
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
			ev.Encode()
		}
	})

	// Decode
	enc := ev.Encode()
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			decodeSubscription(enc)
		}
	})

}
