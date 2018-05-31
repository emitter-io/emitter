/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
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

	"github.com/emitter-io/emitter/collection"
	"github.com/emitter-io/emitter/message"
	"github.com/stretchr/testify/assert"
)

func TestEncodeSubscriptionState(t *testing.T) {
	defer restoreClock(collection.Now)

	setClock(0)
	state := (*subscriptionState)(&collection.LWWSet{
		Set: collection.LWWState{"A": {AddTime: 10, DelTime: 50}},
	})

	// Encode
	enc := state.Encode()[0]
	assert.Equal(t, []byte{0x1, 0x1, 0x41, 0x14, 0x64}, enc)

	// Decode
	dec, err := decodeSubscriptionState(enc)
	assert.NoError(t, err)
	assert.Equal(t, state, dec)
}

func TestEncodeSubscriptionEvent(t *testing.T) {
	defer restoreClock(collection.Now)

	setClock(0)
	ev := SubscriptionEvent{
		Ssid: message.Ssid{1, 2, 3, 4, 5},
		Peer: 657,
		Conn: 12456,
	}

	// Encode
	enc := ev.Encode()

	// Decode
	dec, err := decodeSubscriptionEvent(enc)
	assert.NoError(t, err)
	assert.Equal(t, ev, dec)
}

// RestoreClock restores the clock time
func restoreClock(clk func() int64) {
	collection.Now = clk
}

// SetClock sets the clock time for testing
func setClock(t int64) {
	collection.Now = func() int64 { return t }
	println("clock set to", collection.Now())
}
