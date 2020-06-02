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
	"regexp"

	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service"
)

var (
	shortcut = regexp.MustCompile("^[a-zA-Z0-9]{1,2}$")
)

type keygen interface {
	DecryptKey(string) (security.Key, error)
}

// Service represents a self-introspection service.
type Service struct{}

// New creates a new service.
func New() *Service {
	return new(Service)
}

// OnRequest handles a request to create a link.
func (s *Service) OnRequest(c service.Conn, payload []byte) (service.Response, bool) {
	links := make(map[string]string)
	for k, v := range c.Links() {
		links[k] = security.ParseChannel([]byte(v)).SafeString()
	}

	return &Response{
		ID:    c.ID(),
		Links: links,
	}, true
}
