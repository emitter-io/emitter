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
		return ErrBadRequest
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
		return ErrBadRequest
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
	if channel.ChannelType == security.ChannelInvalid {
		return ErrBadRequest
	}

	// Publish should only have static channel strings
	if channel.ChannelType != security.ChannelStatic {
		return ErrForbidden
	}

	// Is this a special api request?
	if TryProcessAPIRequest(c, channel, payload) {
		return nil
	}

	// Attempt to parse the key
	key, err := c.service.Cipher.DecryptKey(channel.Key)
	if err != nil {
		return ErrUnauthorized
	}

	// Has the key expired?
	if key.IsExpired() {
		return ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract := c.service.ContractProvider.Get(key.Contract())
	if contract == nil {
		return ErrNotFound
	}

	// Validate the contract
	if !contract.Validate(key) {
		return ErrUnauthorized
	}

	// Check if the key has the permission to write here
	if !key.HasPermission(security.AllowWrite) {
		return ErrUnauthorized
	}

	// Check if the key has the permission for the required channel

	kt := key.Target()
	ct := channel.Target()
	if kt != 0 && kt != ct {
		return ErrUnauthorized
	}

	// Do we have a TTL with the message?
	_, hasTTL := channel.TTL()

	// In case of ttl, check the key provides the permission to store (soft permission)
	if hasTTL && !key.HasPermission(security.AllowStore) {
		//ttl = 0
	}

	/*
		// Only call into the storage service if necessary
		if (ttl > 0 && Services.Storage != null)
		{
			// If we have a storage service, store the message
			Services.Storage.AppendAsync(contractId, ssid, ttl, message);
		}
	*/
	// Iterate through all subscribers and send them the message
	ssid := NewSsid(key.Contract(), channel)
	for _, subscriber := range c.service.subscriptions.Lookup(ssid) {
		subscriber.Send(ssid, channel.Channel, payload)
	}

	return nil
}
