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

package cluster

import (
	"testing"

	"github.com/emitter-io/emitter/internal/crdt"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/mesh"
)

func TestEncodeSubscriptionState(t *testing.T) {
	defer restoreClock(crdt.Now)

	setClock(0)
	state := (*subscriptionState)(&crdt.LWWSet{
		Set: crdt.LWWState{"A": {AddTime: 10, DelTime: 50}},
	})

	// Encode
	enc := state.Encode()[0]
	assert.Equal(t, []byte{0x7, 0x18, 0x1, 0x1, 0x41, 0x1, 0x14, 0x1, 0x64}, enc)

	// Decode
	dec, err := decodeSubscriptionState(enc)
	assert.NoError(t, err)
	assert.Equal(t, state, dec)
}

func TestEncodeSubscriptionEvent(t *testing.T) {
	defer restoreClock(crdt.Now)

	setClock(0)
	ev := SubscriptionEvent{
		Ssid:    message.Ssid{1, 2, 3, 4, 5},
		Peer:    657,
		Conn:    12456,
		User:    "hello",
		Channel: []byte("a/b/c/d/e/"),
	}

	// Encode
	enc := ev.Encode()
	assert.Equal(t, 27, len(enc))

	// Decode
	dec, err := decodeSubscriptionEvent(enc)
	assert.NoError(t, err)
	assert.Equal(t, ev, dec)
}

// RestoreClock restores the clock time
func restoreClock(clk func() int64) {
	crdt.Now = clk
}

// SetClock sets the clock time for testing
func setClock(t int64) {
	crdt.Now = func() int64 { return t }
}

func TestEncodeSubscriptionState_Range(t *testing.T) {
	defer restoreClock(crdt.Now)

	setClock(0)
	state := newSubscriptionState()

	for i := 1; i <= 10; i++ {
		ev := SubscriptionEvent{Ssid: message.Ssid{1}, Peer: mesh.PeerName(i % 3), Conn: 777}
		setClock(int64(i))
		state.Add(ev.Encode())
	}

	// Must have 3 keys alive
	setClock(int64(20))
	assert.Equal(t, 3, countAdded(state))

	// Must have 2 keys alive after removal
	setClock(int64(21))

	state.Range(mesh.PeerName(1), func(ev SubscriptionEvent) {
		state.Remove(ev.Encode())
	})
	assert.Equal(t, 2, countAdded(state))
}

func countAdded(state *subscriptionState) (added int) {
	for _, v := range state.Clone() {
		if v.IsAdded() {
			added++
		}
	}
	return
}

// Benchmark_SubscriptionEvent/encode-8         	 4270450	       280 ns/op	     112 B/op	       2 allocs/op
// Benchmark_SubscriptionEvent/decode-8         	 6030177	       199 ns/op	     768 B/op	       3 allocs/op
func Benchmark_SubscriptionEvent(b *testing.B) {
	ev := SubscriptionEvent{
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
	enc := []byte(ev.Encode())
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			decodeSubscriptionState(enc)
		}
	})

}
