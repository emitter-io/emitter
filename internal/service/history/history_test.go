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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	assert.True(t, true)
	// ssid := message.Ssid{1, 3238259379, 500706888, 1027807523}
	// store := storage.NewInMemory(nil)
	// store.Configure(nil)
	// trie := message.NewTrie()
	// auth := &fake.Authorizer{
	// 	Contract: 1,
	// }

	// // Create new service
	// s := New(auth, store)
	// c := &fake.Conn{}

	// // Store a message
	// for i := 0; i < 10; i++ {
	// 	store.Store(&message.Message{
	// 		ID:      message.NewID(ssid),
	// 		Channel: []byte("test/"),
	// 		Payload: []byte("hello"),
	// 		TTL:     30,
	// 	})
	// }

	// request :=
	// err := s.OnRequest(c, )
	// assert.Equal(t, tc.success, err == nil)
	// assert.Equal(t, tc.expectLoaded, len(c.Outgoing))
	// assert.Equal(t, tc.expectCount, trie.Count())

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
