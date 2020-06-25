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

package link

import (
	"encoding/json"
	"testing"

	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/stretchr/testify/assert"
)

func TestLink(t *testing.T) {
	tests := []struct {
		contract int
		request  *Request
		success  bool
	}{
		{request: nil},
		{request: &Request{}},
		{
			contract: 1,
			success:  false,
			request: &Request{
				Name:    "a",
				Channel: "a/b/c/",
			},
		},
		{
			contract: 1,
			success:  true,
			request: &Request{
				Name:    "a",
				Key:     "key",
				Channel: "a/b/c/",
			},
		},
		{
			contract: 1,
			success:  true,
			request: &Request{
				Name:      "a",
				Subscribe: true,
				Key:       "key",
				Channel:   "a/b/c/",
			},
		},
	}

	for _, tc := range tests {
		pubsub := new(fake.PubSub)
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
		s := New(auth, pubsub)
		c := new(fake.Conn)
		_, ok := s.OnRequest(c, b)
		assert.Equal(t, tc.success, ok)

		// Assert the successful result
		if tc.request != nil && tc.success {
			assert.LessOrEqual(t, 5, len(c.GetLink([]byte(tc.request.Name))))

			// If requested to subscribe, make sure we have it
			if tc.request.Subscribe {
				assert.Equal(t, 1, pubsub.Trie.Count())
			}
		}
	}
}
