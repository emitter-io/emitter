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
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/security/hash"
	"github.com/kelindar/binary"
)

const (
	requestKeygen   = 548658350  // hash("keygen")
	requestPresence = 3869262148 // hash("presence")
	requestLink     = 2667034312 // hash("link")
	requestMe       = 2539734036 // hash("me")
)

var (
	shortcut = regexp.MustCompile("^[a-zA-Z0-9]{1,2}$")
)

// ------------------------------------------------------------------------------------

// Authorize attempts to authorize a channel with its key
func (c *Conn) authorize(channel *security.Channel, permission uint8) (contract.Contract, security.Key, bool) {

	// Attempt to parse the key
	key, err := c.service.Cipher.DecryptKey(channel.Key)
	if err != nil || key.IsExpired() {
		return nil, nil, false
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := c.service.contracts.Get(key.Contract())
	if !contractFound || !contract.Validate(key) || !key.HasPermission(permission) || !key.ValidateChannel(channel) {
		return nil, nil, false
	}

	// Return the contract and the key
	return contract, key, true
}

// ------------------------------------------------------------------------------------

// onConnect handles the connection authorization
func (c *Conn) onConnect(packet *mqtt.Connect) bool {
	c.username = string(packet.Username)
	return true
}

// ------------------------------------------------------------------------------------

// OnSubscribe is a handler for MQTT Subscribe events.
func (c *Conn) onSubscribe(mqttTopic []byte) *Error {

	// Parse the channel
	channel := security.ParseChannel(mqttTopic)
	if channel.ChannelType == security.ChannelInvalid {
		return ErrBadRequest
	}

	// Check the authorization and permissions
	contract, key, allowed := c.authorize(channel, security.AllowRead)
	if !allowed {
		return ErrUnauthorized
	}

	// Subscribe the client to the channel
	ssid := message.NewSsid(key.Contract(), channel.Query)
	c.Subscribe(ssid, channel.Channel)

	// Use limit = 1 if not specified, otherwise use the limit option. The limit now
	// defaults to one as per MQTT spec we always need to send retained messages.
	limit := int64(1)
	if v, ok := channel.Last(); ok {
		limit = v
	}

	// Check if the key has a load permission (also applies for retained)
	if key.HasPermission(security.AllowLoad) {
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

	// Check the authorization and permissions
	contract, key, allowed := c.authorize(channel, security.AllowRead)
	if !allowed {
		return ErrUnauthorized
	}

	// Unsubscribe the client from the channel
	ssid := message.NewSsid(key.Contract(), channel.Query)
	c.Unsubscribe(ssid, channel.Channel)
	c.track(contract)
	return nil
}

// ------------------------------------------------------------------------------------

// OnPublish is a handler for MQTT Publish events.
func (c *Conn) onPublish(packet *mqtt.Publish) *Error {
	mqttTopic := packet.Topic
	if len(mqttTopic) <= 2 && c.links != nil {
		mqttTopic = []byte(c.links[string(mqttTopic)])
	}

	// Make sure we have a valid channel
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
		c.onEmitterRequest(channel, packet.Payload, packet.MessageID)
		return nil
	}

	// Check the authorization and permissions
	contract, key, allowed := c.authorize(channel, security.AllowWrite)
	if !allowed {
		return ErrUnauthorized
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
		c.service.storage.Store(msg)
	}

	// Check whether an exclude me option was set (i.e.: 'me=0')
	var exclude string
	if channel.Exclude() {
		exclude = c.ID()
	}

	// Iterate through all subscribers and send them the message
	size := c.service.publish(msg, exclude)

	// Write the monitoring information
	c.track(contract)
	contract.Stats().AddIngress(int64(len(packet.Payload)))
	contract.Stats().AddEgress(size)
	return nil
}

// ------------------------------------------------------------------------------------

// onEmitterRequest processes an emitter request.
func (c *Conn) onEmitterRequest(channel *security.Channel, payload []byte, requestID uint16) (ok bool) {
	var resp response
	defer func() {
		c.sendResponse(channel.String(), resp, requestID)
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
	case requestLink:
		resp, ok = c.onLink(payload)
		return
	default:
		return
	}
}

// ------------------------------------------------------------------------------------

// onLink handles a request to create a link.
func (c *Conn) onLink(payload []byte) (response, bool) {
	var request linkRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		return ErrBadRequest, false
	}

	// Check whether the name is a valid shortcut name
	if !shortcut.Match([]byte(request.Name)) {
		return ErrLinkInvalid, false
	}

	// Make the channel from the request or try to make a private one
	channel := security.MakeChannel(request.Key, request.Channel)
	if request.Private {
		channel = c.makePrivateChannel(request.Key, request.Channel)
	}

	// Ensures that the channel requested is valid
	if channel == nil || channel.ChannelType == security.ChannelInvalid {
		return ErrBadRequest, false
	}

	// Create the link with the name and set the full channel to it
	c.links[request.Name] = channel.String()

	// If an auto-subscribe was requested and the key has read permissions, subscribe
	if _, key, allowed := c.authorize(channel, security.AllowRead); allowed && request.Subscribe {
		c.Subscribe(message.NewSsid(key.Contract(), channel.Query), channel.Channel)
	}

	return &linkResponse{
		Status:  200,
		Name:    request.Name,
		Channel: channel.SafeString(),
	}, true
}

// makePrivateChannel creates a private channel and an appropriate key.
func (c *Conn) makePrivateChannel(chanKey, chanName string) *security.Channel {
	channel := security.MakeChannel(chanKey, chanName)
	if channel.ChannelType != security.ChannelStatic {
		return nil
	}

	// Make sure we can actually extend it
	_, key, allowed := c.authorize(channel, security.AllowExtend)
	if !allowed {
		return nil
	}

	// Create a new key for the private link
	target := fmt.Sprintf("%s%s/", channel.Channel, c.ID())
	if err := key.SetTarget(target); err != nil {
		return nil
	}

	// Encrypt the key for storing
	encryptedKey, err := c.service.Cipher.EncryptKey(key)
	if err != nil {
		return nil
	}

	// Create the private channel
	channel.Channel = []byte(target)
	channel.Query = append(channel.Query, hash.Of([]byte(c.ID())))
	channel.Key = []byte(encryptedKey)
	return channel
}

// ------------------------------------------------------------------------------------

// OnMe is a handler that returns information to the connection.
func (c *Conn) onMe() (response, bool) {
	links := make(map[string]string)
	for k, v := range c.links {
		links[k] = security.ParseChannel([]byte(v)).SafeString()
	}

	return &meResponse{
		ID:    c.ID(),
		Links: links,
	}, true
}

// ------------------------------------------------------------------------------------

// onKeyGen processes a keygen request.
func (c *Conn) onKeyGen(payload []byte) (response, bool) {
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
	for _, subscriber := range s.subscriptions.Lookup(ssid, nil) {
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
func (c *Conn) onPresence(payload []byte) (response, bool) {
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
	ssid := message.NewSsid(key.Contract(), channel.Query)

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
