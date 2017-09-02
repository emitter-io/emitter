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
	"strings"
	"time"

	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/emitter-io/emitter/encoding"
	"github.com/emitter-io/emitter/logging"
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

	// Has the key expired?
	if key.IsExpired() {
		return ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract := c.service.Contracts.Get(key.Contract())
	if contract == nil {
		return ErrNotFound
	}

	// Validate the contract
	if !contract.Validate(key) {
		return ErrUnauthorized
	}

	// Check if the key has the permission to read from here
	if !key.HasPermission(security.AllowRead) {
		return ErrUnauthorized
	}

	// Check if the key has the permission for the required channel
	if key.Target() != 0 && key.Target() != channel.Target() {
		return ErrUnauthorized
	}

	// Subscribe the client to the channel
	ssid := subscription.NewSsid(key.Contract(), channel)
	c.Subscribe(ssid, channel.Channel)

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

	// Has the key expired?
	if key.IsExpired() {
		return ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract := c.service.Contracts.Get(key.Contract())
	if contract == nil {
		return ErrNotFound
	}

	// Validate the contract
	if !contract.Validate(key) {
		return ErrUnauthorized
	}

	// Check if the key has the permission to read from here
	if !key.HasPermission(security.AllowRead) {
		return ErrUnauthorized
	}

	// Check if the key has the permission for the required channel
	if key.Target() != 0 && key.Target() != channel.Target() {
		return ErrUnauthorized
	}

	// Unsubscribe the client from the channel
	ssid := subscription.NewSsid(key.Contract(), channel)
	c.Unsubscribe(ssid, channel.Channel)

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
		c.onEmitterRequest(channel, payload)
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
	contract := c.service.Contracts.Get(key.Contract())
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

	// In case of ttl, check the key provides the permission to store (soft permission)
	if _, ok := channel.TTL(); ok && key.HasPermission(security.AllowStore) {

		/*
			// Only call into the storage service if necessary
			if (Services.Storage != null)
			{
				// If we have a storage service, store the message
				Services.Storage.AppendAsync(contractId, ssid, ttl, message);
			}
		*/
	}

	// Iterate through all subscribers and send them the message
	size := c.service.publish(subscription.NewSsid(key.Contract(), channel), channel.Channel, payload)

	// Write the stats
	contract.Stats().AddIngress(int64(len(payload)))
	contract.Stats().AddEgress(size)
	return nil
}

// ------------------------------------------------------------------------------------

// onEmitterRequest processes an emitter request.
func (c *Conn) onEmitterRequest(channel *security.Channel, payload []byte) (ok bool) {
	var resp interface{}
	defer func() {
		if b, err := json.Marshal(resp); err == nil {
			c.Send(nil, []byte("emitter/"+string(channel.Channel)), b)
		}
	}()

	// Make sure we have a query
	resp = ErrNotFound
	if len(channel.Query) < 1 {
		return
	}

	switch channel.Query[0] {
	case requestKeygen:
		resp, ok = c.onKeyGen(payload)
		return
	case requestPresence:
		resp, ok = c.onPresence(payload)
		return
	default:
		return
	}
}

// ------------------------------------------------------------------------------------

// onKeyGen processes a keygen request.
func (c *Conn) onKeyGen(payload []byte) (interface{}, bool) {
	// Deserialize the payload.
	message := keyGenRequest{}
	if err := json.Unmarshal(payload, &message); err != nil {
		return ErrBadRequest, false
	}

	// Attempt to parse the key, this should be a master key
	masterKey, err := c.service.Cipher.DecryptKey([]byte(message.Key))
	if err != nil || !masterKey.IsMaster() || masterKey.IsExpired() {
		return ErrUnauthorized, false
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract := c.service.Contracts.Get(masterKey.Contract())
	if contract == nil {
		return ErrNotFound, false
	}

	// Validate the contract
	if !contract.Validate(masterKey) {
		return ErrUnauthorized, false
	}

	// Use the cipher to generate the key
	key, err := c.service.Cipher.GenerateKey(masterKey, message.Channel, message.access(), message.expires(), -1)
	if err != nil {
		return ErrServerError, false
	}

	// Success, return the response
	return &keyGenResponse{
		Status:  200,
		Key:     key,
		Channel: message.Channel,
	}, true

}

// ------------------------------------------------------------------------------------

// onPresenceQuery handles an incoming presence query.
func (s *Service) onPresenceQuery(queryType string, payload []byte) ([]byte, bool) {
	if queryType != "presence" {
		return nil, false
	}

	// Decode the request
	var target subscription.Ssid
	if err := encoding.Decode(payload, &target); err != nil {
		return nil, false
	}

	logging.LogTarget("query", queryType+" query received", target)

	// Send back the response
	if b, err := encoding.Encode(s.lookupPresence(target)); err == nil {
		return b, true
	}
	return nil, false
}

// lookupPresence performs a subscriptions lookup and returns a presence information.
func (s *Service) lookupPresence(ssid subscription.Ssid) []presenceInfo {
	resp := make([]presenceInfo, 0, 4)
	for _, subscriber := range s.subscriptions.Lookup(ssid) {
		if conn, ok := subscriber.(*Conn); ok {
			resp = append(resp, presenceInfo{
				ID:       conn.ID(),
				Username: conn.username,
			})
		}
	}
	return resp
}

// ------------------------------------------------------------------------------------

// onKeyGen processes a keygen request.
func (c *Conn) onPresence(payload []byte) (interface{}, bool) {
	// Deserialize the payload.
	message := presenceRequest{}
	if err := json.Unmarshal(payload, &message); err != nil {
		return ErrBadRequest, false
	}

	// Attempt to parse the key, this should be a master key
	key, err := c.service.Cipher.DecryptKey([]byte(message.Key))
	if err != nil || !key.HasPermission(security.AllowPresence) || key.IsExpired() {
		return ErrUnauthorized, false
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract := c.service.Contracts.Get(key.Contract())
	if contract == nil {
		return ErrNotFound, false
	}

	// Validate the contract
	if !contract.Validate(key) {
		return ErrUnauthorized, false
	}

	// Ensure we have trailing slash
	if !strings.HasSuffix(message.Channel, "/") {
		message.Channel = message.Channel + "/"
	}

	// Parse the channel
	channel := security.ParseChannel([]byte("emitter/" + message.Channel))
	if channel.ChannelType == security.ChannelInvalid {
		return ErrBadRequest, false
	}

	// Create the ssid for the presence
	ssid := subscription.NewSsid(key.Contract(), channel)

	// Check if the client is interested in subscribing/unsubscribing from changes.
	if message.Changes {
		c.Subscribe(subscription.NewSsidForPresence(ssid), nil)
	} else {
		c.Unsubscribe(subscription.NewSsidForPresence(ssid), nil)
	}

	// If we requested a status, populate the slice via scatter/gather.
	now := time.Now().UTC().Unix()
	who := make([]presenceInfo, 0, 4)
	if message.Status {

		// Gather local presence first
		who = append(who, c.service.lookupPresence(ssid)...)

		// Issue the presence query to the cluster
		if req, err := encoding.Encode(ssid); err == nil {
			if awaiter, err := c.service.Query("presence", req); err == nil {

				// Wait for all presence updates to come back (or a deadline)
				for _, resp := range awaiter.Gather(1000 * time.Millisecond) {
					info := []presenceInfo{}
					if err := encoding.Decode(resp, &info); err == nil {
						//logging.LogTarget("query", "response gathered", info)
						who = append(who, info...)
					}
				}
			}
		}
	}

	return &presenceResponse{
		Time:    now,
		Event:   presenceStatusEvent,
		Channel: message.Channel,
		Who:     who,
	}, true
}
