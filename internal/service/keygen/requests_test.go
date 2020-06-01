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

package keygen

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_access(t *testing.T) {
	r := &Request{
		Type: "rwslpex",
	}

	assert.Zero(t, r.expires().Unix())
	assert.Equal(t, 254, int(r.access()))
}

func Test_expires(t *testing.T) {
	req := &Request{
		TTL: 20,
	}

	assert.Less(t, time.Now().Unix(), req.expires().Unix())

	res := new(Response)
	res.ForRequest(1)
	assert.Equal(t, 1, int(res.Request))
}
