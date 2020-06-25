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
	"time"

	"github.com/emitter-io/emitter/internal/security"
)

// Request represents a key generation request.
type Request struct {
	Key     string `json:"key"`     // The master key to use.
	Channel string `json:"channel"` // The channel to create a key for.
	Type    string `json:"type"`    // The permission set.
	TTL     int32  `json:"ttl"`     // The TTL of the key.
}

// expires returns the requested expiration time
func (m *Request) expires() time.Time {
	if m.TTL == 0 {
		return time.Unix(0, 0)
	}

	return time.Now().Add(time.Duration(m.TTL) * time.Second).UTC()
}

// access returns the requested level of access
func (m *Request) access() uint8 {
	required := security.AllowNone
	for i := 0; i < len(m.Type); i++ {
		switch c := m.Type[i]; c {
		case 'r':
			required |= security.AllowRead
		case 'w':
			required |= security.AllowWrite
		case 's':
			required |= security.AllowStore
		case 'l':
			required |= security.AllowLoad
		case 'p':
			required |= security.AllowPresence
		case 'e':
			required |= security.AllowExtend
		case 'x':
			required |= security.AllowExecute
		}
	}

	return required
}

// ------------------------------------------------------------------------------------

// Response represents a key generation response
type Response struct {
	Request uint16 `json:"req,omitempty"`
	Status  int    `json:"status"`
	Key     string `json:"key"`
	Channel string `json:"channel"`
}

// ForRequest sets the request ID in the response for matching
func (r *Response) ForRequest(id uint16) {
	r.Request = id
}
