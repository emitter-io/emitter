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

package me

import (
	"testing"

	"github.com/emitter-io/emitter/internal/service/fake"
	"github.com/stretchr/testify/assert"
)

func TestMe(t *testing.T) {
	s := New()
	c := &fake.Conn{
		Shortcuts: map[string]string{
			"a": "test",
		},
	}

	r, ok := s.OnRequest(c, nil)
	assert.Equal(t, true, ok)

	resp := r.(*Response)
	assert.Contains(t, resp.Links, "a")

}
