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

// Request represents a key ban request.
type Request struct {
	Secret string `json:"secret"` // The master key to use.
	Target string `json:"target"` // The target key to ban.
	Banned bool   `json:"banned"` // Whether the target should be banned or not.
}

// ------------------------------------------------------------------------------------

// Response represents a key ban response.
type Response struct {
	Request uint16 `json:"req,omitempty"`
	Status  int    `json:"status"` // The status of the response
	Banned  bool   `json:"banned"` // Whether the target should be banned or not.
}

// ForRequest sets the request ID in the response for matching
func (r *Response) ForRequest(id uint16) {
	r.Request = id
}
