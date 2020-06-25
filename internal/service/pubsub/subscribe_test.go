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
	"errors"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/stretchr/testify/assert"
)

func TestPubSub_Subscribe(t *testing.T) {
	ssid := message.Ssid{1, 3238259379, 500706888, 1027807523}
	tests := []struct {
		contract     int    // The contract ID
		topic        string // The subscribe topic
		extraPerm    uint8  // Extra key permission
		disabled     bool   // Is the connection disabled?
		expectLoaded int    // How many messages were loaded?
		expectCount  int    // How many subscribers now?
		success      bool   // Success or failure?
	}{
		{ // Bad request
			success: false,
			topic:   "",
		},
		{ // Unauthorized
			success: false,
			topic:   "key/a/b/c/",
		},
		{ // Unauthorized, extend
			success:   false,
			contract:  1,
			extraPerm: security.AllowExtend,
			topic:     "key/a/b/c/",
		},
		{ // Happy Path, simple
			success:     true,
			contract:    1,
			expectCount: 1,
			topic:       "key/a/b/c/",
		},
		{ // Disabled
			success:     true,
			contract:    1,
			disabled:    true,
			expectCount: 0,
			topic:       "key/a/b/c/",
		},
		{ // Unauthorized retained
			success:      true,
			contract:     1,
			expectCount:  1,
			expectLoaded: 0,
			topic:        "key/a/b/c/?last=10",
		},
		{ // Happy Path, retained
			success:      true,
			contract:     1,
			extraPerm:    security.AllowLoad,
			expectCount:  1,
			expectLoaded: 7,
			topic:        "key/a/b/c/?last=7",
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

		// Create new service
		s := New(auth, store, notify, trie)
		c := &fake.Conn{
			Disabled: tc.disabled,
		}

		// Store a message
		for i := 0; i < 10; i++ {
			store.Store(&message.Message{
				ID:      message.NewID(ssid),
				Channel: []byte("a/b/c/"),
				Payload: []byte("hello"),
				TTL:     30,
			})
		}

		err := s.OnSubscribe(c, []byte(tc.topic))
		assert.Equal(t, tc.success, err == nil)
		assert.Equal(t, tc.expectLoaded, len(c.Outgoing))
		assert.Equal(t, tc.expectCount, trie.Count())
	}
}

func TestPubSub_Subscribe_Buggy(t *testing.T) {
	tests := []struct {
		contract     int    // The contract ID
		topic        string // The subscribe topic
		extraPerm    uint8  // Extra key permission
		disabled     bool   // Is the connection disabled?
		expectLoaded int    // How many messages were stored?
		expectCount  int    // How many messages were published?
		success      bool   // Success or failure?
	}{
		{ // Buggy
			success:      false,
			contract:     1,
			extraPerm:    security.AllowLoad,
			expectCount:  1,
			expectLoaded: 0,
			topic:        "key/a/b/c/?last=7",
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
		s := New(auth, new(buggyStore), new(fake.Notifier), trie)
		c := &fake.Conn{
			Disabled: tc.disabled,
		}

		err := s.OnSubscribe(c, []byte(tc.topic))
		assert.Equal(t, tc.success, err == nil)
		assert.Equal(t, tc.expectLoaded, len(c.Outgoing))
		assert.Equal(t, tc.expectCount, trie.Count())
	}
}

// ------------------------------------------------------------------------------------

// Noop implements Storage contract.
var _ storage.Storage = new(buggyStore)

// Noop represents a storage which does nothing.
type buggyStore struct{}

// Name returns the name of the provider.
func (s *buggyStore) Name() string {
	return "noop"
}

func (s *buggyStore) Configure(config map[string]interface{}) error {
	return errors.New("not working")
}

func (s *buggyStore) Store(m *message.Message) error {
	return errors.New("not working")
}

func (s *buggyStore) Query(ssid message.Ssid, from, until time.Time, limit int) (message.Frame, error) {
	return nil, errors.New("not working")
}

func (s *buggyStore) Close() error {
	return errors.New("not working")
}
