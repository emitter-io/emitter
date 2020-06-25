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
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
)

// OnLastWill publishes a last will event of the subscriber.
func (s *Service) OnLastWill(sub message.Subscriber, ev *event.Connection) bool {
	if ev == nil || !ev.WillFlag {
		return false
	}

	// Make sure we have a valid channel
	channel := security.ParseChannel(ev.WillTopic)
	if channel.ChannelType != security.ChannelStatic {
		return false
	}

	// Check the authorization and permissions
	contract, key, allowed := s.auth.Authorize(channel, security.AllowWrite)
	if !allowed || key.HasPermission(security.AllowExtend) {
		return false
	}

	// Create a new message
	msg := message.New(
		message.NewSsid(key.Contract(), channel.Query),
		channel.Channel,
		ev.WillMessage,
	)

	// If a user have specified a retain flag, retain with a default TTL
	if ev.WillRetain {
		msg.TTL = message.RetainedTTL
	}

	// Store the message if needed
	if msg.Stored() && key.HasPermission(security.AllowStore) {
		s.store.Store(msg)
	}

	// Iterate through all subscribers and send them the message
	size := s.Publish(msg, nil)

	// Write the monitoring information
	contract.Stats().AddIngress(int64(len(ev.WillMessage)))
	contract.Stats().AddEgress(size)
	return true
}
