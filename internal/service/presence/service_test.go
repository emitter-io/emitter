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

package presence

import (
	"testing"

	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/kelindar/binary"
	"github.com/kelindar/binary/nocopy"
	"github.com/stretchr/testify/assert"
)

func TestPresence_OnSurvey(t *testing.T) {
	ssid := message.Ssid{1, 3238259379, 500706888, 1027807523}
	pubsub := new(fake.PubSub)
	pubsub.Subscribe(new(fake.Conn), &event.Subscription{
		Peer:    2,
		Conn:    5,
		Ssid:    ssid,
		Channel: nocopy.Bytes("a/b/c/"),
	})

	s := New(&fake.Authorizer{
		Contract: 1,
		Success:  true,
	}, pubsub, new(fake.Surveyor), pubsub.Trie)

	// Bad query
	{
		_, ok := s.OnSurvey("xxx", nil)
		assert.False(t, ok)
	}

	// Bad request
	{
		_, ok := s.OnSurvey("presence", nil)
		assert.False(t, ok)
	}

	{
		b, _ := binary.Marshal(ssid)
		r, ok := s.OnSurvey("presence", b)

		var out []Info
		err := binary.Unmarshal(r, &out)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, 1, len(out))
	}
}

func TestPresence_Notify(t *testing.T) {
	ssid := message.Ssid{1, 3238259379, 500706888, 1027807523}
	pubsub := new(fake.PubSub)
	pubsub.Subscribe(new(fake.Conn), &event.Subscription{
		Peer:    2,
		Conn:    5,
		Ssid:    ssid,
		Channel: nocopy.Bytes("a/b/c/"),
	})

	s := New(&fake.Authorizer{
		Contract: 1,
		Success:  true,
	}, pubsub, new(fake.Surveyor), pubsub.Trie)

	s.Notify(EventTypeSubscribe, &event.Subscription{
		Ssid: ssid,
	}, nil)
	s.pollPresenceChange()
	assert.Equal(t, 1, s.trie.Count())

}
