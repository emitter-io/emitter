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

package keyban

import (
	"encoding/json"
	"testing"

	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/stretchr/testify/assert"
)

func TestKeyBan(t *testing.T) {
	tests := []struct {
		contract1 int
		contract2 int
		perms     uint8
		request   *Request
		initial   string
		expected  string
		success   bool
	}{
		{request: nil},
		{request: &Request{}},
		{
			contract1: 1,
			contract2: 1,
			perms:     security.AllowMaster,
			success:   true,
			expected:  "b",
			request: &Request{
				Secret: "a",
				Target: "b",
				Banned: true,
			},
		},
		{
			contract1: 1,
			contract2: 1,
			perms:     security.AllowMaster,
			success:   true,
			initial:   "b",
			request: &Request{
				Secret: "a",
				Target: "b",
				Banned: false,
			},
		},
		{
			contract1: 1,
			contract2: 2,
			request: &Request{
				Secret: "a",
				Target: "b",
				Banned: true,
			},
		},
	}

	for _, tc := range tests {
		repl := new(fake.Replicator)
		s := New(&fake.Authorizer{
			Contract: uint32(tc.contract1),
			Success:  tc.contract1 != 0,
		}, &fake.Decryptor{
			Contract:    uint32(tc.contract2),
			Permissions: tc.perms,
		}, repl)

		// Fill the replicator
		initial := event.Ban(tc.initial)
		repl.Notify(&initial, tc.initial != "")

		// Prepare the request
		b, _ := json.Marshal(tc.request)
		if tc.request == nil {
			b = []byte("invalid")
		}

		// Issue a request
		_, ok := s.OnRequest(nil, b)
		assert.Equal(t, tc.success, ok)

		// Make sure we have the key if expected
		expected := event.Ban(tc.expected)
		assert.Equal(t, tc.expected != "", repl.Contains(&expected))
	}
}
