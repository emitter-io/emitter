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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/kelindar/binary"
	"github.com/kelindar/binary/nocopy"
	"github.com/stretchr/testify/assert"
)

func TestPresence_OnRequest(t *testing.T) {
	bTrue, bFalse := true, false
	tests := []struct {
		contract     int
		request      *Request
		expectStatus int
		expectSubs   int
		success      bool
	}{
		{request: nil},
		{request: &Request{}},
		{
			contract: 0,
			success:  false,
			request: &Request{
				Key:     "key",
				Channel: "a/b/c/",
				Status:  true,
			},
		},
		{
			contract:   1,
			success:    true,
			expectSubs: 1,
			request: &Request{
				Key:     "key",
				Channel: "a/b/c/",
				Status:  false,
			},
		},
		{
			contract:     1,
			success:      true,
			expectStatus: 5,
			expectSubs:   1,
			request: &Request{
				Key:     "key",
				Channel: "a/b/c/",
				Status:  true,
			},
		},
		{
			contract:     1,
			success:      true,
			expectStatus: 5,
			expectSubs:   2,
			request: &Request{
				Key:     "key",
				Channel: "a/b/c/",
				Status:  true,
				Changes: &bTrue,
			},
		},
		{
			contract:     1,
			success:      true,
			expectStatus: 5,
			expectSubs:   1,
			request: &Request{
				Key:     "key",
				Channel: "a/b/c/",
				Status:  true,
				Changes: &bFalse,
			},
		},
	}

	for _, tc := range tests {
		users, _ := binary.Marshal([]Info{{ID: "user1"}, {ID: "user2"}})
		survey := &fake.Surveyor{
			Resp: [][]byte{users, users},
		}

		pubsub := new(fake.PubSub)
		pubsub.Subscribe(new(fake.Conn), &event.Subscription{
			Peer:    2,
			Conn:    5,
			Ssid:    message.Ssid{1, 3238259379, 500706888, 1027807523},
			Channel: nocopy.Bytes("a/b/c/"),
		})

		auth := &fake.Authorizer{
			Contract: uint32(tc.contract),
			Success:  tc.contract != 0,
		}

		// Prepare the request
		b, _ := json.Marshal(tc.request)
		if tc.request == nil {
			b = []byte("invalid")
		} else {
			auth.Target = tc.request.Channel
		}

		// Issue a request
		s := New(auth, pubsub, survey, pubsub.Trie)
		defer s.Close()

		c := new(fake.Conn)
		r, ok := s.OnRequest(c, b)
		assert.Equal(t, tc.success, ok)

		if resp, ok := r.(*Response); ok {
			assert.Equal(t, tc.expectStatus, len(resp.Who))
		}

		if tc.success {
			assert.Equal(t, tc.expectSubs, pubsub.Trie.Count())
		}
	}
}

func TestPresence_OnHTTP(t *testing.T) {
	tests := []struct {
		contract     int
		method       string
		request      *Request
		expectStatus int
		expectSubs   int
		code         int
	}{
		{
			method: "GET",
			code:   404,
		},
		{
			method: "POST",
			code:   400,
		},
		{
			method:  "POST",
			request: &Request{},
			code:    400,
		},
		{
			method:   "POST",
			contract: 0,
			code:     401,
			request: &Request{
				Key:     "key",
				Channel: "a/b/c/",
			},
		},
		{
			method:       "POST",
			contract:     1,
			code:         200,
			expectStatus: 5,
			expectSubs:   1,
			request: &Request{
				Key:     "key",
				Channel: "a/b/c/",
			},
		},
	}

	for _, tc := range tests {
		users, _ := binary.Marshal([]Info{{ID: "user1"}, {ID: "user2"}})
		survey := &fake.Surveyor{
			Resp: [][]byte{users, users},
		}

		pubsub := new(fake.PubSub)
		pubsub.Subscribe(new(fake.Conn), &event.Subscription{
			Peer:    2,
			Conn:    5,
			Ssid:    message.Ssid{1, 3238259379, 500706888, 1027807523},
			Channel: nocopy.Bytes("a/b/c/"),
		})

		auth := &fake.Authorizer{
			Contract: uint32(tc.contract),
			Success:  tc.contract != 0,
		}

		// Prepare the request
		b, _ := json.Marshal(tc.request)
		if tc.request == nil {
			b = []byte("invalid")
		} else {
			auth.Target = tc.request.Channel
		}

		// Issue a request
		s := New(auth, pubsub, survey, pubsub.Trie)
		defer s.Close()

		req, _ := http.NewRequest(tc.method, "/presence", bytes.NewBuffer(b))
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(s.OnHTTP)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, tc.code, rr.Code)
		if tc.code == 200 {
			var resp Response
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectStatus, len(resp.Who))
		}
	}
}
