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
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/emitter-io/emitter/internal/service/me"
	"github.com/kelindar/binary/nocopy"
	"github.com/stretchr/testify/assert"
)

func TestPubSub_Publish(t *testing.T) {
	ssid := message.Ssid{1, 3238259379, 500706888, 1027807523}
	tests := []struct {
		contract     int           // The contract ID
		request      *mqtt.Publish // The publish request
		extraPerm    uint8         // Extra key permission
		expectStored int           // How many messages were stored?
		expectCount  int           // How many messages were published?
		success      bool          // Success or failure?
	}{
		{ // Bad request
			success: false,
			request: &mqtt.Publish{},
		},
		{ // Non-static
			success: false,
			request: &mqtt.Publish{
				Topic: []byte("key/a/+/c/"),
			},
		},
		{ // Unauthorized
			success: false,
			request: &mqtt.Publish{
				Topic: []byte("key/a/b/c/"),
			},
		},
		{ // Unauthorized, extend
			contract:  1,
			success:   false,
			extraPerm: security.AllowExtend,
			request: &mqtt.Publish{
				Topic: []byte("key/a/b/c/"),
			},
		},
		{ // Happy Path, Simple
			contract:    1,
			expectCount: 1,
			success:     true,
			request: &mqtt.Publish{
				Topic: []byte("key/a/b/c/"),
			},
		},
		{ // Happy Path, No Echo
			contract:    1,
			expectCount: 0,
			success:     true,
			request: &mqtt.Publish{
				Topic: []byte("key/a/b/c/?me=0"),
			},
		},
		{ // // Happy Path, Retained
			contract:     1,
			success:      true,
			extraPerm:    security.AllowStore,
			expectStored: 1,
			expectCount:  1,
			request: &mqtt.Publish{
				Topic: []byte("key/a/b/c/?ttl=30"),
				Header: mqtt.Header{
					Retain: true,
				},
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

		c := new(fake.Conn)
		err := s.OnPublish(c, tc.request)
		assert.Equal(t, tc.success, err == nil)
		assert.Equal(t, tc.expectCount, len(sub.Outgoing))

		// Query the storage
		{
			msgs, err := store.Query(ssid, time.Unix(0, 0), time.Now(), 100)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectStored, len(msgs))
		}
	}
}

func TestPubSub_Request(t *testing.T) {
	tests := []struct {
		contract int           // The contract ID
		request  *mqtt.Publish // The publish request
		success  bool          // Success or failure?
	}{
		{ // Happy Path, Simple
			contract: 1,
			success:  true,
			request: &mqtt.Publish{
				Topic: []byte("emitter/me/"),
			},
		},
		{ // Happy Path, No Handler
			contract: 1,
			success:  true,
			request: &mqtt.Publish{
				Topic: []byte("emitter/xxx/"),
			},
		},
	}

	for _, tc := range tests {
		trie := message.NewTrie()
		auth := &fake.Authorizer{
			Contract: uint32(tc.contract),
			Success:  tc.contract != 0,
		}

		// Issue a request
		s := New(auth, storage.NewNoop(), new(fake.Notifier), trie)
		s.Handle("me", me.New().OnRequest)

		c := new(fake.Conn)
		err := s.OnPublish(c, tc.request)
		assert.Equal(t, tc.success, err == nil)

	}
}
