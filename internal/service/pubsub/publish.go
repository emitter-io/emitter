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
	"encoding/json"
	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service"
)

// Publish publishes a message to everyone and returns the number of outgoing bytes written.
func (s *Service) Publish(m *message.Message, filter func(message.Subscriber) bool) (n int64) {
	size := m.Size()
	for _, subscriber := range s.trie.Lookup(m.Ssid(), filter) {
		subscriber.Send(m)
		if subscriber.Type() == message.SubscriberDirect {
			n += size
		}
	}
	return
}

// OnPublish is a handler for MQTT Publish events.
func (s *Service) OnPublish(c service.Conn, packet *mqtt.Publish) *errors.Error {
	mqttTopic := c.GetLink(packet.Topic)

	// Make sure we have a valid channel
	channel := security.ParseChannel(mqttTopic)
	if channel.ChannelType == security.ChannelInvalid {
		return errors.ErrBadRequest
	}

	// Publish should only have static channel strings
	if channel.ChannelType != security.ChannelStatic {
		return errors.ErrForbidden
	}

	// Check whether the key is 'emitter' which means it's an API request
	if len(channel.Key) == 7 && string(channel.Key) == "emitter" {
		s.onEmitterRequest(c, channel, packet.Payload, packet.MessageID)
		return nil
	}

	// Check the authorization and permissions
	contract, key, allowed := s.auth.Authorize(channel, security.AllowWrite)
	if !allowed {
		return errors.ErrUnauthorized
	}

	// Keys which are supposed to be extended should not be used for publishing
	if key.HasPermission(security.AllowExtend) {
		return errors.ErrUnauthorizedExt
	}

	// Create a new message
	msg := message.New(
		message.NewSsid(key.Contract(), channel.Query),
		channel.Channel,
		packet.Payload,
	)

	// If a user have specified a retain flag, retain with a default TTL
	if packet.Header.Retain {
		msg.TTL = message.RetainedTTL
	}

	// If a user have specified a TTL, use that value
	if ttl, ok := channel.TTL(); ok && ttl > 0 {
		msg.TTL = uint32(ttl)
	}

	// Store the message if needed
	if msg.Stored() && key.HasPermission(security.AllowStore) {
		s.store.Store(msg)
	}

	// Check whether an exclude me option was set (i.e.: 'me=0')
	var exclude string
	if channel.Exclude() {
		exclude = c.ID()
	}

	// Iterate through all subscribers and send them the message
	size := s.Publish(msg, func(s message.Subscriber) bool {
		return s.ID() != exclude
	})

	// Write the monitoring information
	c.Track(contract)
	contract.Stats().AddIngress(int64(len(packet.Payload)))
	contract.Stats().AddEgress(size)
	return nil
}

// onEmitterRequest processes an emitter request.
func (s *Service) onEmitterRequest(c service.Conn, channel *security.Channel, payload []byte, requestID uint16) (ok bool) {
	var resp service.Response
	defer func() {
		if resp != nil {
			sendResponse(c, channel.String(), resp, requestID)
		}
	}()

	// Make sure we have a query
	resp = errors.ErrNotFound
	if len(channel.Query) == 1 {
		if handle, ok := s.handlers[channel.Query[0]]; ok {
			resp, ok = handle(c, payload)
		}
	}
	return
}

// Sends a response back to the client.
func sendResponse(c service.Conn, topic string, resp service.Response, requestID uint16) {
	switch m := resp.(type) {
	case *errors.Error:
		cpy := m.Copy()
		cpy.ForRequest(requestID)
		resp = cpy
	default:
		m.ForRequest(requestID)
	}

	if b, err := json.Marshal(resp); err == nil {
		c.Send(&message.Message{
			Channel: []byte(topic),
			Payload: b,
		})
	}
	return
}
