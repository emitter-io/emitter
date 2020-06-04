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
	"time"

	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/kelindar/binary/nocopy"
	"github.com/stretchr/testify/assert"
)

func TestPubSub_LastWill(t *testing.T) {
	ssid := message.Ssid{1, 3238259379, 500706888, 1027807523}
	tests := []struct {
		contract     int               // The contract ID
		event        *event.Connection // The event
		extraPerm    uint8             // Extra key permission
		expectStored int               // How many messages were stored?
		expectCount  int               // How many messages were published?
		success      bool              // Success or failure?
	}{
		{ // Nil request
			success: false,
		},
		{ // Bad request
			success: false,
			event: &event.Connection{
				Peer: 1,
				Conn: 2,
			},
		},
		{ // Bad request, no channel
			success: false,
			event: &event.Connection{
				Peer:     1,
				Conn:     2,
				WillFlag: true,
			},
		},
		{ // Unauthorized
			success: false,
			event: &event.Connection{
				Peer:      1,
				Conn:      2,
				WillFlag:  true,
				WillTopic: []byte("key/a/b/c/"),
			},
		},
		{ // Happy Path
			success:     true,
			contract:    1,
			expectCount: 1,
			event: &event.Connection{
				Peer:      1,
				Conn:      2,
				WillFlag:  true,
				WillTopic: []byte("key/a/b/c/"),
			},
		},
		{ // Not retained
			success:      true,
			contract:     1,
			expectCount:  1,
			expectStored: 0,
			event: &event.Connection{
				Peer:       1,
				Conn:       2,
				WillFlag:   true,
				WillTopic:  []byte("key/a/b/c/"),
				WillRetain: true,
			},
		},
		{ // Retained
			success:      true,
			contract:     1,
			expectCount:  1,
			expectStored: 1,
			extraPerm:    security.AllowStore,
			event: &event.Connection{
				Peer:       1,
				Conn:       2,
				WillFlag:   true,
				WillTopic:  []byte("key/a/b/c/"),
				WillRetain: true,
			},
		},
	}

	for _, tc := range tests {
		store := storage.NewInMemory(nil)
		store.Configure(nil)
		trie := message.NewTrie()
		notify := new(fake.Notifier)
		auth := &fake.Authorizer{
			Contract:  uint32(tc.contract),
			Success:   tc.contract != 0,
			ExtraPerm: tc.extraPerm,
		}

		// Issue a request
		s := New(auth, store, notify, trie)
		sub := new(fake.Conn)
		s.Subscribe(sub, &event.Subscription{
			Peer:    2,
			Conn:    5,
			Ssid:    ssid,
			Channel: nocopy.Bytes("a/b/c/"),
		})

		assert.Equal(t, tc.success, s.OnLastWill(sub, tc.event))
		assert.Equal(t, tc.expectCount, len(sub.Outgoing))

		// Query the storage
		{
			msgs, err := store.Query(ssid, time.Unix(0, 0), time.Now(), 100)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectStored, len(msgs))
		}
	}
}
