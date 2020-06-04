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

package pubsub

import (
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/security/hash"
	"github.com/emitter-io/emitter/internal/service"
)

// Service represents a publish service.
type Service struct {
	auth     service.Authorizer         // The authorizer to use.
	store    storage.Storage            // The storage provider to use.
	notifier service.Notifier           // The notifier to use.
	trie     *message.Trie              // The subscription matching trie.
	handlers map[uint32]service.Handler // The emitter request handlers.
}

// New creates a new publisher service.
func New(auth service.Authorizer, store storage.Storage, notifier service.Notifier, trie *message.Trie) *Service {
	return &Service{
		auth:     auth,
		store:    store,
		notifier: notifier,
		trie:     trie,
		handlers: make(map[uint32]service.Handler),
	}
}

// Handle adds a handler for an "emitter/..." request
func (s *Service) Handle(request string, handler service.Handler) {
	s.handlers[hash.OfString(request)] = handler
}
