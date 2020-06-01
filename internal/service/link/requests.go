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

// Request represents a link generation request.
type Request struct {
	Name      string `json:"name"`      // The name of the shortcut, max 2 characters.
	Key       string `json:"key"`       // The key for the channel.
	Channel   string `json:"channel"`   // The channel name for the shortcut.
	Subscribe bool   `json:"subscribe"` // Specifies whether the broker should auto-subscribe.
}

// ------------------------------------------------------------------------------------

// Response represents a link generation response.
type Response struct {
	Request uint16 `json:"req,omitempty"`     // The corresponding request ID.
	Status  int    `json:"status"`            // The status of the response.
	Name    string `json:"name,omitempty"`    // The name of the shortcut, max 2 characters.
	Channel string `json:"channel,omitempty"` // The channel which was registered.
}

// ForRequest sets the request ID in the response for matching
func (r *Response) ForRequest(id uint16) {
	r.Request = id
}
