/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
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

package broker

import (
	"github.com/emitter-io/emitter/security"
)

// EventError represents an event code which provides a more de.
type EventError struct {
	code int
	msg  string
}

// Error implements error interface.
func (e *EventError) Error() string { return e.msg }

// Represents a set of errors used in the handlers.
var (
	ErrBadRequest      = &EventError{code: 400, msg: "The request was invalid or cannot be otherwise served."}
	ErrUnauthorized    = &EventError{code: 401, msg: "The security key provided is not authorized to perform this operation."}
	ErrPaymentRequired = &EventError{code: 402, msg: "The request can not be served, as the payment is required to proceed."}
	ErrForbidden       = &EventError{code: 403, msg: "The request is understood, but it has been refused or access is not allowed."}
	ErrNotFound        = &EventError{code: 404, msg: "The resource requested does not exist."}
	ErrServerError     = &EventError{code: 500, msg: "An unexpected condition was encountered and no more specific message is suitable."}
	ErrNotImplemented  = &EventError{code: 501, msg: "The server either does not recognize the request method, or it lacks the ability to fulfill the request."}
)

// OnSubscribe is a handler for MQTT Subscribe events.
func (c *Conn) onSubscribe(mqttTopic []byte) *EventError {

	// Parse the channel
	channel := security.ParseChannel(mqttTopic)
	if channel.ChannelType == security.ChannelInvalid {
		return ErrBadRequest
	}

	// Attempt to parse the key
	key, err := c.service.Cipher.DecryptKey(channel.Key)
	if err != nil {
		// TODO
	}

	// Subscribe the client to the channel
	c.Subscribe(key.Contract(), channel)

	return nil
}

// OnUnsubscribe is a handler for MQTT Unsubscribe events.
func (c *Conn) onUnsubscribe(mqttTopic []byte) *EventError {

	// Parse the channel
	channel := security.ParseChannel(mqttTopic)
	if channel.ChannelType == security.ChannelInvalid {
		return ErrBadRequest
	}

	// Attempt to parse the key
	key, err := c.service.Cipher.DecryptKey(channel.Key)
	if err != nil {
		// TODO
	}

	// Unsubscribe the client from the channel
	ssid := NewSsid(key.Contract(), channel)
	c.Unsubscribe(ssid)

	return nil
}

// OnPublish is a handler for MQTT Publish events.
func (c *Conn) onPublish(mqttTopic []byte, payload []byte) *EventError {

	// Parse the channel
	channel := security.ParseChannel(mqttTopic)
	if channel.ChannelType != security.ChannelStatic {
		return ErrForbidden
	}

	// Is this a special api request?
	/*if TryProcessAPIRequest(channel, payload) {
		return nil
	}*/

	// Attempt to parse the key
	key, err := c.service.Cipher.DecryptKey(channel.Key)
	if err != nil {
		// TODO
	}

	// Has the key expired?

	// Attempt to fetch the contract using the key. Underneath, it's cached.

	// Check if the payment state is valid

	// Validate the contract

	// Check if the key has the permission to write here

	// Check if the key has the permission for the required channel

	// Do we have a TTL with the message?

	// Check if the key has a TTL and also can store (soft permission)

	// Iterate through all subscribers and send them the message
	ssid := NewSsid(key.Contract(), channel)
	for _, subscriber := range c.service.subscriptions.Lookup(ssid) {
		subscriber.Send(channel.Channel, payload)
	}

	return nil
}
