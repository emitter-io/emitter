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

package history

import (
	"encoding/json"
	"testing"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	//assert.True(t, true)
	ssid := message.Ssid{1, 3238259379, 500706888, 1027807523}
	store := storage.NewInMemory(nil)
	store.Configure(nil)
	auth := &fake.Authorizer{
		Success:   true,
		Contract:  uint32(1),
		ExtraPerm: security.AllowLoad,
	}

	request := &Request{
		Key:     "key",
		Channel: "a/b/c/",
	}

	// Prepare the request
	b, _ := json.Marshal(request)
	if request == nil {
		b = []byte("invalid")
	} else {
		auth.Target = request.Channel
	}

	// Create new service
	s := New(auth, store)
	c := &fake.Conn{}

	// Store a message
	for i := 0; i < 1; i++ {
		store.Store(&message.Message{
			ID:      message.NewID(ssid),
			Channel: []byte("a/b/c/?ttl=30"),
			Payload: []byte("hello"),
			TTL:     30,
		})
	}

	// Issue a request
	response, ok := s.OnRequest(c, b)
	println(response)
	assert.Equal(t, true, ok)
}

// ------------------------------------------------------------------------------------

// // Noop implements Storage contract.
// var _ storage.Storage = new(buggyStore)

// // Noop represents a storage which does nothing.
// type buggyStore struct{}

// // Name returns the name of the provider.
// func (s *buggyStore) Name() string {
// 	return "noop"
// }

// func (s *buggyStore) Configure(config map[string]interface{}) error {
// 	return errors.New("not working")
// }

// func (s *buggyStore) Store(m *message.Message) error {
// 	return errors.New("not working")
// }

// func (s *buggyStore) Query(ssid message.Ssid, from, until time.Time, limit int) (message.Frame, error) {
// 	return nil, errors.New("not working")
// }

// func (s *buggyStore) Close() error {
// 	return errors.New("not working")
// }
