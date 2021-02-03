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
	"bytes"
	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service"
	"github.com/kelindar/binary/nocopy"
)

// Subscribe subscribes to a channel.
func (s *Service) Subscribe(sub message.Subscriber, ev *event.Subscription) bool {
	if conn, ok := sub.(service.Conn); ok && !conn.CanSubscribe(ev.Ssid, ev.Channel) {
		return false
	}

	// Add the subscription to the trie
	s.trie.Subscribe(ev.Ssid, sub)

	// Broadcast direct subscriptions
	s.notifier.NotifySubscribe(sub, ev)
	return true
}

// OnSubscribe is a handler for MQTT Subscribe events.
func (s *Service) OnSubscribe(c service.Conn, mqttTopic []byte) *errors.Error {

	// compatibility with paho.mqtt.golang
	// https://github.com/eclipse/paho.mqtt.golang/blob/master/topic.go#L78
	// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Toc385349376
	mqttTopic = bytes.ReplaceAll(mqttTopic, []byte("#"), []byte("#/"))
	mqttTopic = bytes.ReplaceAll(mqttTopic, []byte("//"), []byte("/"))

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

	// Subscribe the client to the channel
	ssid := message.NewSsid(key.Contract(), channel.Query)
	s.Subscribe(c, &event.Subscription{
		Conn:    c.LocalID(),
		User:    nocopy.String(c.Username()),
		Ssid:    ssid,
		Channel: channel.Channel,
	})

	// Use limit = 1 if not specified, otherwise use the limit option. The limit now
	// defaults to one as per MQTT spec we always need to send retained messages.
	limit := int64(1)
	if v, ok := channel.Last(); ok {
		limit = v
	}

	// Check if the key has a load permission (also applies for retained)
	if key.HasPermission(security.AllowLoad) {
		t0, t1 := channel.Window() // Get the window
		msgs, err := s.store.Query(ssid, t0, t1, int(limit))
		if err != nil {
			logging.LogError("conn", "query last messages", err)
			return errors.ErrServerError
		}

		// Range over the messages in the channel and forward them
		for _, m := range msgs {
			msg := m // Copy message
			c.Send(&msg)
		}
	}

	// Write the stats
	c.Track(contract)
	return nil
}
