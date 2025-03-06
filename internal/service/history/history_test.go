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
	"crypto/rand"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/network/mqtt"
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

func TestLargeMessage(t *testing.T) {
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

	// Store 1 long message
	// Keep in mind the message will be composed of the ID and the channel size on top of the payload.
	// So mqttMaxMessageSize is really smaller than the actual message size.
	randomBytes := make([]byte, mqtt.MaxMessageSize)
	rand.Read(randomBytes)
	firstSSID := message.NewID(ssid)
	store.Store(&message.Message{
		ID:      firstSSID,
		Channel: []byte("a/b/c/"),
		Payload: randomBytes,
		TTL:     30,
	})

	reqBytes, _ := json.Marshal(request)

	// Issue the same request
	response, ok := service.OnRequest(connection, reqBytes)
	// The request should have succeeded and returned a response.
	assert.Equal(t, true, ok)
	// The response should have returned the last message as per MQTT spec.
	assert.Equal(t, 0, len(response.(*Response).Messages))
}

// ONLY PASSES BECAUSE OF THE BUG, THERE IS ONLY ONE SERVER SO NO GATHER
// match.Limit(limit) only limits based on the number of messages not the size of the frame
/*func (s *SSD) Query(ssid message.Ssid, from, until time.Time, startFromID message.ID, limit int) (message.Frame, error) {

	// Construct a query and lookup locally first
	query := newLookupQuery(ssid, from, until, startFromID, limit)
	match := s.lookup(query)

	// Issue the message survey to the cluster
	if req, err := binary.Marshal(query); err == nil && s.survey != nil {
		if awaiter, err := s.survey.Query("ssdstore", req); err == nil {

			// Wait for all presence updates to come back (or a deadline)
			for _, resp := range awaiter.Gather(2000 * time.Millisecond) {
				if frame, err := message.DecodeFrame(resp); err == nil {
					match = append(match, frame...)
				}
			}
		}
	}

	match.Limit(limit)
	return match, nil
}*/
func TestSumOfTwoExceedMaxSize(t *testing.T) {
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
	fmt.Println(int(mqtt.MaxMessageSize - len(firstSSID) - len("test/") - 1)) // KEYSIZE???
	//randomBytes := make([]byte, int(mqtt.MaxMessageSize-len(firstSSID)-len("a/b/c/")-1)) // BUG: MaxMessageSize represents the maximum size of the payload, but the message is composed of the ID, the channel size and the payload.
	randomBytes := make([]byte, int(mqtt.MaxMessageSize))
	rand.Read(randomBytes)
	err := store.Store(&message.Message{
		ID:      firstSSID,
		Channel: []byte("a/b/c/"),
		Payload: randomBytes,
		TTL:     30,
	})
	assert.NoError(t, err)
	store.Store(&message.Message{
		ID:      message.NewID(ssid),
		Channel: []byte("a/b/c/"),
		Payload: randomBytes,
		TTL:     30,
	})
	reqBytes, _ := json.Marshal(request)

	request.Channel = "key/a/b/c/?last=2"
	reqBytes, _ = json.Marshal(request)
	response, ok := service.OnRequest(connection, reqBytes)
	// The request should have succeeded and returned a response.
	assert.Equal(t, true, ok)
	// The response should have returned the last 2 messages.
	assert.Equal(t, 1, len(response.(*Response).Messages))
}
