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
	"github.com/weaveworks/mesh"
	"strconv"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/event/crdt"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/kelindar/binary/nocopy"
	"github.com/stretchr/testify/assert"
)

func restoreClock(clk func() int64) {
	crdt.Now = clk
}

// SetClock sets the clock time for testing
func setClock(t int64) {
	crdt.Now = func() int64 { return t }
}

// Benchmark_State/contains-8         	11538494	       100 ns/op	      16 B/op	       1 allocs/op
func Benchmark_State(b *testing.B) {
	state := NewState(":memory:")
	for i := 1; i <= 20000; i++ {
		ev := Ban(strconv.Itoa(i))
		setClock(int64(i))
		state.Add(&ev)
	}

	// Encode
	target := Ban("10000")
	state.Has(&target)
	time.Sleep(10 * time.Millisecond)
	b.Run("contains", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			state.Has(&target)
		}
	})
}

func TestEncodeSubscriptionState(t *testing.T) {
	defer restoreClock(crdt.Now)
	for _, tc := range []struct {
		dir string
	}{
		{dir: ":memory:"},
		{dir: ""},
	} {
		state := NewState(tc.dir)
		setClock(10)
		ev := &Subscription{
			Channel: nocopy.Bytes("A"),
		}
		state.Add(ev)
		assert.True(t, state.Has(ev))

		// Encode / decode
		enc := state.Encode()[0]
		dec, err := DecodeState(enc)
		assert.NoError(t, err)
		dec.Subscriptions(func(k *Subscription, v Value) {
			assert.Equal(t, "A", string(k.Channel))
			assert.Equal(t, int64(10), v.AddTime())
			assert.Equal(t, int64(0), v.DelTime())
		})

		state.Close()
	}
}

func TestMergeState(t *testing.T) {
	defer restoreClock(crdt.Now)
	ev := Subscription{
		Channel: nocopy.Bytes("A"),
	}

	// Add to state 1
	setClock(20)
	state1 := NewState("")
	state1.Add(&ev)

	// Del from state 2
	setClock(50)
	state2 := NewState("")
	state2.Del(&ev)

	// Merge
	delta := state1.Merge(state2)
	assert.Equal(t, state2, delta)

	state1.Subscriptions(func(_ *Subscription, v Value) {
		assert.Equal(t, int64(20), v.AddTime())
		assert.Equal(t, int64(50), v.DelTime())
	})

	state2.Subscriptions(func(_ *Subscription, v Value) {
		assert.Equal(t, int64(50), v.DelTime())
	})

	// Merge with zero delta
	state3 := NewState("")
	state3.Del(&ev)
	delta = state3.Merge(state2)
	assert.Nil(t, delta)
}

func TestSubscriptions(t *testing.T) {
	defer restoreClock(crdt.Now)

	setClock(0)
	state := NewState(":memory:")
	defer state.Close()

	for i := 1; i <= 10; i++ {
		ev := Subscription{Ssid: message.Ssid{1}, Peer: uint64(i) % 3, Conn: 777}
		setClock(int64(i))
		state.Add(&ev)
	}

	// Must have 3 keys alive
	setClock(int64(20))
	assert.Equal(t, 3, countAdded(state))

	// Must have 2 keys alive after removal
	setClock(int64(21))
	state.SubscriptionsOf(1, func(ev *Subscription) {
		state.Del(ev)
	})
	assert.Equal(t, 2, countAdded(state))

	// Count all of the subscriptions (alive or dead)
	count := 0
	state.Subscriptions(func(ev *Subscription, _ Value) {
		count++
	})
	assert.Equal(t, 3, count)
}

func TestConnections(t *testing.T) {
	defer restoreClock(crdt.Now)

	setClock(0)
	state := NewState(":memory:")
	defer state.Close()

	for i := 1; i <= 10; i++ {
		ev := Connection{Peer: uint64(i) % 3, Conn: 777}
		setClock(int64(i))
		state.Add(&ev)
	}

	count := 0
	state.ConnectionsOf(mesh.PeerName(2), func(*Connection) {
		count++
	})
	assert.Equal(t, 1, count)
}

func countAdded(state *State) (added int) {
	set := state.subsets[typeSub]
	set.Range(nil, false, func(_ string, v Value) bool {
		added++
		return true
	})
	return
}
