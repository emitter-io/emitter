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

// TestHistory tests the history service.
func TestHistory(t *testing.T) {
	ssid := message.Ssid{1, 3238259379, 500706888, 1027807523}
	store := storage.NewInMemory(nil)
	store.Configure(nil)
	auth := &fake.Authorizer{
		Success:   true,
		Contract:  uint32(1),
		ExtraPerm: security.AllowLoad,
	}
	// Create new service
	service := New(auth, store)
	connection := &fake.Conn{}

	// The most basic request, on an empty store.
	request := &Request{
		Key:     "key",
		Channel: "key/a/b/c/",
	}

	// Store 2 messages
	firstSSID := message.NewID(ssid)
	store.Store(&message.Message{
		ID:      firstSSID,
		Channel: []byte("a/b/c/"),
		Payload: []byte("hello"),
		TTL:     30,
	})
	store.Store(&message.Message{
		ID:      message.NewID(ssid),
		Channel: []byte("a/b/c/"),
		Payload: []byte("hello"),
		TTL:     30,
	})
	reqBytes, _ := json.Marshal(request)

	// Issue the same request
	response, ok := service.OnRequest(connection, reqBytes)
	// The request should have succeeded and returned a response.
	assert.Equal(t, true, ok)
	// The response should have returned the last message as per MQTT spec.
	assert.Equal(t, 1, len(response.(*Response).Messages))

	store.Store(&message.Message{
		ID:      message.NewID(ssid),
		Channel: []byte("a/b/c/"),
		Payload: []byte("hello"),
		TTL:     30,
	})
	request.Channel = "key/a/b/c/?last=2"
	reqBytes, _ = json.Marshal(request)
	response, ok = service.OnRequest(connection, reqBytes)
	// The request should have succeeded and returned a response.
	assert.Equal(t, true, ok)
	// The response should have returned the last 2 messages.
	assert.Equal(t, 2, len(response.(*Response).Messages))
}
