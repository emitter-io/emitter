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
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service"
	"github.com/kelindar/binary"
)

var (
	shortcut = regexp.MustCompile("^[a-zA-Z0-9]{1,2}$")
)

// Service represents a key blacklisting service.
type Service struct {
	connection service.Conn
	auth       service.Authorizer // The authorizer to use.
	keygen     service.Decryptor  // The key generator to use.
	cluster    service.Replicator // The cluster service to use.
	queue      chan *Notification // The channel for keyban notifications.
	context    context.Context    // The context for the service.
	cancel     context.CancelFunc // The cancellation function.
	//subscriptions *message.Trie      // The subscription matching trie.
}

// New creates a new key blacklisting service.
func New(auth service.Authorizer, keygen service.Decryptor, cluster service.Replicator /*, subscriptions *message.Trie*/) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{
		context: ctx,
		cancel:  cancel,
		auth:    auth,
		keygen:  keygen,
		cluster: cluster,
		//subscriptions: subscriptions,
		queue: make(chan *Notification, 100),
	}

	s.pollKeybanChange()
	return s
}

func (s *Service) pollKeybanChange() {
	go func() {
		for {
			select {
			case <-s.context.Done():
				return
			case notif := <-s.queue:
				// Depending on the flag, ban or unban the key
				bannedKey := event.Ban(notif.Key)
				switch {
				case notif.Banned && !s.cluster.Contains(&bannedKey):
					s.connection.Ban()
					s.connection.Close()
					s.cluster.Notify(&bannedKey, true)
				case !notif.Banned && s.cluster.Contains(&bannedKey):
					s.cluster.Notify(&bannedKey, false)
				}
			}
		}
	}()
}

// OnRequest handles a request to ban or unban a key.
func (s *Service) OnRequest(c service.Conn, payload []byte) (service.Response, bool) {
	var message Request
	if err := json.Unmarshal(payload, &message); err != nil {
		return errors.ErrBadRequest, false
	}

	// Decrypt the secret key and make sure it's not expired and is a master key
	_, secretKey, ok := s.auth.Authorize(security.ParseChannel(
		binary.ToBytes(fmt.Sprintf("%s/emitter/", message.Secret)),
	), security.AllowMaster)
	if !ok || secretKey.IsExpired() || !secretKey.IsMaster() {
		return errors.ErrUnauthorized, false
	}

	// Make sure the target key is for the same contract
	targetKey, err := s.keygen.DecryptKey(message.Target)
	if err != nil || targetKey.Contract() != secretKey.Contract() {
		return errors.ErrUnauthorized, false
	}

	// Depending on the flag, ban or unban the key
	bannedKey := event.Ban(message.Target)
	switch {
	case message.Banned && !s.cluster.Contains(&bannedKey):
		c.Ban()
		c.Close()
		s.cluster.Notify(&bannedKey, true)
	case !message.Banned && s.cluster.Contains(&bannedKey):
		s.cluster.Notify(&bannedKey, false)
	}

	// Success, return the response
	return &Response{
		Status: 200,
		Banned: message.Banned,
	}, true
}
