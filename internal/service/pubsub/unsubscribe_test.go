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

package pubsub

import (
	"testing"

	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/kelindar/binary/nocopy"
	"github.com/stretchr/testify/assert"
)

func TestPubSub_Unsubscribe(t *testing.T) {
	ssid := message.Ssid{1, 3238259379, 500706888, 1027807523}
	tests := []struct {
		contract    int    // The contract ID
		topic       string // The topic
		extraPerm   uint8  // Extra key permission
		disabled    bool   // Is the connection disabled?
		expectCount int    // How many subscribers now?
		success     bool   // Success or failure?
	}{
		{ // Bad request
			success:     false,
			expectCount: 10,
			topic:       "",
		},
		{ // Unauthorized
			success:     false,
			expectCount: 10,
			topic:       "key/a/b/c/",
		},
		{ // Unauthorized, extend
			success:     false,
			contract:    1,
			expectCount: 10,
			extraPerm:   security.AllowExtend,
			topic:       "key/a/b/c/",
		},
		{ // Happy Path, simple
			success:     true,
			contract:    1,
			expectCount: 9,
			topic:       "key/a/b/c/",
		},
		{ // Disabled
			success:     true,
			contract:    1,
			disabled:    true,
			expectCount: 10,
			topic:       "key/a/b/c/",
		},
	}

	for _, tc := range tests {
		trie := message.NewTrie()
		auth := &fake.Authorizer{
			Contract:  uint32(tc.contract),
			Success:   tc.contract != 0,
			ExtraPerm: tc.extraPerm,
		}

		// Create new service
		s := New(auth, storage.NewNoop(), new(fake.Notifier), trie)

		// Register few subscribers
		for i := 0; i < 10; i++ {
			assert.True(t, s.Subscribe(&fake.Conn{
				ConnID: i,
			}, &event.Subscription{
				Peer:    2,
				Conn:    security.ID(i),
				Ssid:    ssid,
				Channel: nocopy.Bytes("a/b/c/"),
			}))
		}

		err := s.OnUnsubscribe(&fake.Conn{
			ConnID:   5,
			Disabled: tc.disabled,
		}, []byte(tc.topic))
		assert.Equal(t, tc.success, err == nil)
		assert.Equal(t, tc.expectCount, trie.Count())
	}
}
