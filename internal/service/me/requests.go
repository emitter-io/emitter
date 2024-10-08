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

// Response represents a response for the 'me' request.
type Response struct {
	Request uint16            `json:"req,omitempty"`   // The corresponding request ID.
	Status  int               `json:"status"`          // The status of the response
	ID      string            `json:"id"`              // The private ID of the connection.
	Links   map[string]string `json:"links,omitempty"` // The set of pre-defined channels.
}

// ForRequest sets the request ID in the response for matching
func (r *Response) ForRequest(id uint16) {
	r.Request = id
}
