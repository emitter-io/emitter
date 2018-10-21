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

	"github.com/emitter-io/emitter/message"
	"github.com/emitter-io/emitter/provider/logging"
	"github.com/emitter-io/emitter/security"
	"github.com/kelindar/binary"
)

const (
	requestKeygen   = 548658350
	requestPresence = 3869262148
	requestMe       = 2539734036
)

// ------------------------------------------------------------------------------------

// OnSubscribe is a handler for MQTT Subscribe events.
func (c *Conn) onSubscribe(mqttTopic []byte) *Error {

	// Parse the channel
	channel := security.ParseChannel(mqttTopic)
	if channel.ChannelType == security.ChannelInvalid {
		return ErrBadRequest
	}

	// Attempt to parse the key
	key, err := c.service.Cipher.DecryptKey(channel.Key)
	if err != nil || key.IsExpired() {
		return ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := c.service.contracts.Get(key.Contract())
	if !contractFound {
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
	if !key.ValidateChannel(channel) {
		return ErrUnauthorized
	}

	// Subscribe the client to the channel
	ssid := message.NewSsid(key.Contract(), channel)
	c.Subscribe(ssid, channel.Channel)

	// In case of ttl, check the key provides the permission to store (soft permission)
	if limit, ok := channel.Last(); ok && key.HasPermission(security.AllowLoad) {
		t0, t1 := channel.Window() // Get the window
		msgs, err := c.service.storage.Query(ssid, t0, t1, int(limit))
		if err != nil {
			logging.LogError("conn", "query last messages", err)
			return ErrServerError
		}

		// Range over the messages in the channel and forward them
		for _, m := range msgs {
			msg := m // Copy message
			c.Send(&msg)
		}
	}

	// Write the stats
	c.track(contract)
	return nil
}

// ------------------------------------------------------------------------------------

// OnUnsubscribe is a handler for MQTT Unsubscribe events.
func (c *Conn) onUnsubscribe(mqttTopic []byte) *Error {

	// Parse the channel
	channel := security.ParseChannel(mqttTopic)
	if channel.ChannelType == security.ChannelInvalid {
		return ErrBadRequest
	}

	// Attempt to parse the key
	key, err := c.service.Cipher.DecryptKey(channel.Key)
	if err != nil || key.IsExpired() {
		return ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := c.service.contracts.Get(key.Contract())
	if !contractFound {
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
	if !key.ValidateChannel(channel) {
		return ErrUnauthorized
	}

	// Unsubscribe the client from the channel
	ssid := message.NewSsid(key.Contract(), channel)
	c.Unsubscribe(ssid, channel.Channel)
	c.track(contract)
	return nil
}

// ------------------------------------------------------------------------------------

// OnPublish is a handler for MQTT Publish events.
func (c *Conn) onPublish(mqttTopic []byte, payload []byte) *Error {

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
	if err != nil || key.IsExpired() {
		return ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := c.service.contracts.Get(key.Contract())
	if !contractFound {
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
	if !key.ValidateChannel(channel) {
		return ErrUnauthorized
	}

	// Create a new message
	msg := message.New(
		message.NewSsid(key.Contract(), channel),
		channel.Channel,
		payload,
	)

	// In case of ttl, check the key provides the permission to store (soft permission)
	if ttl, ok := channel.TTL(); ok && key.HasPermission(security.AllowStore) {
		msg.TTL = uint32(ttl) // Add the TTL to the message
		c.service.storage.Store(msg)
	}

	// Iterate through all subscribers and send them the message
	size := c.service.publish(msg)

	// Write the monitoring information
	c.track(contract)
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
			c.Send(&message.Message{
				Channel: []byte("emitter/" + string(channel.Channel)), // TODO: reduce allocations
				Payload: b,
			})
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
	case requestMe:
		resp, ok = c.onMe()
		return
	default:
		return
	}
}

// ------------------------------------------------------------------------------------

// OnMe is a handler that returns information to the connection.
func (c *Conn) onMe() (interface{}, bool) {
	// Success, return the response
	return &meResponse{
		ID: c.ID(),
	}, true
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
	contract, contractFound := c.service.contracts.Get(masterKey.Contract())
	if !contractFound {
		return ErrNotFound, false
	}

	// Validate the contract
	if !contract.Validate(masterKey) {
		return ErrUnauthorized, false
	}

	// Use the cipher to generate the key
	key, err := c.service.Cipher.GenerateKey(masterKey, message.Channel, message.access(), message.expires(), -1)
	if err != nil {
		switch err {
		case security.ErrTargetInvalid:
			return ErrTargetInvalid, false
		case security.ErrTargetTooLong:
			return ErrTargetTooLong, false
		default:
			return ErrServerError, false
		}
	}

	// Success, return the response
	return &keyGenResponse{
		Status:  200,
		Key:     key,
		Channel: message.Channel,
	}, true
}

// ------------------------------------------------------------------------------------

// OnSurvey handles an incoming presence query.
func (s *Service) OnSurvey(queryType string, payload []byte) ([]byte, bool) {
	if queryType != "presence" {
		return nil, false
	}

	// Decode the request
	var target message.Ssid
	if err := binary.Unmarshal(payload, &target); err != nil {
		return nil, false
	}

	logging.LogTarget("query", queryType+" query received", target)

	// Send back the response
	presence, err := binary.Marshal(s.lookupPresence(target))
	return presence, err == nil
}

// lookupPresence performs a subscriptions lookup and returns a presence information.
func (s *Service) lookupPresence(ssid message.Ssid) []presenceInfo {
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

func getClusterPresence(s *Service, ssid message.Ssid) []presenceInfo {
	who := make([]presenceInfo, 0, 4)
	if req, err := binary.Marshal(ssid); err == nil {
		if awaiter, err := s.Survey("presence", req); err == nil {

			// Wait for all presence updates to come back (or a deadline)
			for _, resp := range awaiter.Gather(1000 * time.Millisecond) {
				info := []presenceInfo{}
				if err := binary.Unmarshal(resp, &info); err == nil {
					//logging.LogTarget("query", "response gathered", info)
					who = append(who, info...)
				}
			}
		}
	}
	return who
}

func getLocalPresence(s *Service, ssid message.Ssid) []presenceInfo {
	return s.lookupPresence(ssid)
}

func getAllPresence(s *Service, ssid message.Ssid) []presenceInfo {
	return append(getLocalPresence(s, ssid), getClusterPresence(s, ssid)...)
}

// onPresence processes a presence request.
func (c *Conn) onPresence(payload []byte) (interface{}, bool) {
	// Deserialize the payload.
	msg := presenceRequest{
		Status:  true, // Default: send status info
		Changes: true, // Default: send all changes
	}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return ErrBadRequest, false
	}

	// Attempt to parse the key, this should be a master key
	key, err := c.service.Cipher.DecryptKey([]byte(msg.Key))
	if err != nil || !key.HasPermission(security.AllowPresence) || key.IsExpired() {
		return ErrUnauthorized, false
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := c.service.contracts.Get(key.Contract())
	if !contractFound {
		return ErrNotFound, false
	}

	// Validate the contract
	if !contract.Validate(key) {
		return ErrUnauthorized, false
	}

	// Ensure we have trailing slash
	if !strings.HasSuffix(msg.Channel, "/") {
		msg.Channel = msg.Channel + "/"
	}

	// Parse the channel
	channel := security.ParseChannel([]byte("emitter/" + msg.Channel))
	if channel.ChannelType == security.ChannelInvalid {
		return ErrBadRequest, false
	}

	// Create the ssid for the presence
	ssid := message.NewSsid(key.Contract(), channel)

	// Check if the client is interested in subscribing/unsubscribing from changes.
	if msg.Changes {
		c.Subscribe(message.NewSsidForPresence(ssid), nil)
	} else {
		c.Unsubscribe(message.NewSsidForPresence(ssid), nil)
	}

	// If we requested a status, populate the slice via scatter/gather.
	now := time.Now().UTC().Unix()
	who := make([]presenceInfo, 0, 4)
	if msg.Status {

		// Gather local & cluster presence
		who = append(who, getAllPresence(c.service, ssid)...)
	}

	return &presenceResponse{
		Time:    now,
		Event:   presenceStatusEvent,
		Channel: msg.Channel,
		Who:     who,
	}, true
}
