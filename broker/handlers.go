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
	"encoding/json"

	"github.com/emitter-io/emitter/security"
)

const (
	requestKeygen   = 548658350
	requestPresence = 3869262148
)

// ------------------------------------------------------------------------------------

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

// ------------------------------------------------------------------------------------

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

// ------------------------------------------------------------------------------------

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

	// Check whether the key is 'emitter' which means it's an API request
	if len(channel.Key) == 7 && string(channel.Key) == "emitter" {
		return c.onEmitterRequest(channel, payload)
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
	if key.Target() != 0 && key.Target() != channel.Target() {
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

	// Write the ingress stats
	contract.Stats().AddIngress(int64(len(payload)))

	// Iterate through all subscribers and send them the message
	ssid := NewSsid(key.Contract(), channel)
	for _, subscriber := range c.service.subscriptions.Lookup(ssid) {
		subscriber.Send(ssid, channel.Channel, payload)
	}

	return nil
}

// ------------------------------------------------------------------------------------

// onEmitterRequest processes an emitter request.
func (c *Conn) onEmitterRequest(channel *security.Channel, payload []byte) (err *EventError) {

	switch channel.Query[1] {
	case requestKeygen:
		c.onKeyGen(c.service, channel, payload)
		return nil
	case requestPresence:
		return nil
	default:
		return nil
	}
}

// ------------------------------------------------------------------------------------

// onKeyGen processes a keygen request.
func (c *Conn) onKeyGen(s *Service, channel *security.Channel, payload []byte) (string, error) {
	// Deserialize the payload.
	message := keyGenMessage{}
	if err := json.Unmarshal(payload, &message); err != nil {
		return "", err
	}

	// Attempt to parse the key, this should be a master key
	masterKey, err := s.Cipher.DecryptKey([]byte(message.Key))
	if err != nil {
		return "", err
	}
	if !masterKey.IsMaster() || masterKey.IsExpired() {
		return "", ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract := s.ContractProvider.Get(masterKey.Contract())
	if contract == nil {
		return "", ErrNotFound
	}

	// Validate the contract
	if !contract.Validate(masterKey) {
		return "", ErrUnauthorized
	}

	// Use the cipher to generate the key
	return s.Cipher.GenerateKey(masterKey, message.Channel, message.access(), message.expires())
}
