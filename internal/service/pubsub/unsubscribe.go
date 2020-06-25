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
	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service"
	"github.com/kelindar/binary/nocopy"
)

// Unsubscribe unsubscribes from a channel
func (s *Service) Unsubscribe(sub message.Subscriber, ev *event.Subscription) (ok bool) {
	if conn, ok := sub.(service.Conn); ok && !conn.CanUnsubscribe(ev.Ssid, ev.Channel) {
		return false
	}

	subscribers := s.trie.Lookup(ev.Ssid, nil)
	if ok = subscribers.Contains(sub); ok {
		s.trie.Unsubscribe(ev.Ssid, sub)
	}

	s.notifier.NotifyUnsubscribe(sub, ev)
	return
}

// OnUnsubscribe is a handler for MQTT Unsubscribe events.
func (s *Service) OnUnsubscribe(c service.Conn, mqttTopic []byte) *errors.Error {

	// Parse the channel
	channel := security.ParseChannel(mqttTopic)
	if channel.ChannelType == security.ChannelInvalid {
		return errors.ErrBadRequest
	}

	// Check the authorization and permissions
	contract, key, allowed := s.auth.Authorize(channel, security.AllowRead)
	if !allowed {
		return errors.ErrUnauthorized
	}

	// Keys which are supposed to be extended should not be used for subscribing
	if key.HasPermission(security.AllowExtend) {
		return errors.ErrUnauthorizedExt
	}

	// Unsubscribe the client from the channel
	ssid := message.NewSsid(key.Contract(), channel.Query)
	s.Unsubscribe(c, &event.Subscription{
		Conn:    c.LocalID(),
		User:    nocopy.String(c.Username()),
		Ssid:    ssid,
		Channel: channel.Channel,
	})

	c.Track(contract)
	return nil
}
