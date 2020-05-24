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
	"github.com/kelindar/binary/nocopy"
	"testing"

	"github.com/emitter-io/emitter/internal/event/crdt"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
)

func restoreClock(clk func() int64) {
	crdt.Now = clk
}

// SetClock sets the clock time for testing
func setClock(t int64) {
	crdt.Now = func() int64 { return t }
}

func TestEncodeSubscriptionState(t *testing.T) {
	defer restoreClock(crdt.Now)

	state := NewState()

	setClock(10)
	state.Add(Subscription{
		Channel: nocopy.Bytes("A"),
	})

	// Encode
	enc := state.Encode()[0]
	assert.Equal(t, []byte{0x1, 0x0, 0x1, 0x14, 0x0, 0x6, 0x0, 0x0, 0x0, 0x1, 0x41, 0x0}, enc)

	// Decode
	dec, err := DecodeState(enc)
	assert.NoError(t, err)
	assert.Equal(t, state, dec)
}

func TestMergeState(t *testing.T) {
	defer restoreClock(crdt.Now)
	ev := Subscription{
		Channel: nocopy.Bytes("A"),
	}

	// Add to state 1
	setClock(20)
	state1 := NewState()
	state1.Add(ev)

	// Remove from state 2
	setClock(50)
	state2 := NewState()
	state2.Remove(ev)

	// Merge
	delta := state1.Merge(state2)
	assert.Equal(t, state2, delta)

	state1.Subscriptions(func(_ Subscription, v Time) {
		assert.Equal(t, Time{
			AddTime: 20,
			DelTime: 50,
		}, v)
	})

	state2.Subscriptions(func(_ Subscription, v Time) {
		assert.Equal(t, Time{
			AddTime: 0,
			DelTime: 50,
		}, v)
	})
}

func TestSubscriptions(t *testing.T) {
	defer restoreClock(crdt.Now)

	setClock(0)
	state := NewState()

	for i := 1; i <= 10; i++ {
		ev := Subscription{Ssid: message.Ssid{1}, Peer: uint64(i) % 3, Conn: 777}
		setClock(int64(i))
		state.Add(ev)
	}

	// Must have 3 keys alive
	setClock(int64(20))
	assert.Equal(t, 3, countAdded(state))

	// Must have 2 keys alive after removal
	setClock(int64(21))
	state.SubscriptionsOf(1, func(ev Subscription) {
		state.Remove(ev)
	})
	assert.Equal(t, 2, countAdded(state))

	// Count all of the subscriptions (alive or dead)
	count := 0
	state.Subscriptions(func(ev Subscription, _ Time) {
		count++
	})
	assert.Equal(t, 3, count)
}

func countAdded(state *State) (added int) {
	set := state.m[typeSubscription]
	set.Range(nil, func(_ string, v Time) bool {
		if v.IsAdded() {
			added++
		}
		return true
	})
	return
}
