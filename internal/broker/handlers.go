/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
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
	"github.com/kelindar/binary/nocopy"
	"regexp"
	"strings"
	"time"

	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/kelindar/binary"
)

const (
	requestKeygen   = 548658350  // hash("keygen")
	requestKeyban   = 861724010  // hash("keyban")
	requestPresence = 3869262148 // hash("presence")
	requestLink     = 2667034312 // hash("link")
	requestMe       = 2539734036 // hash("me")
)

var (
	shortcut = regexp.MustCompile("^[a-zA-Z0-9]{1,2}$")
)

// ------------------------------------------------------------------------------------

// onConnect handles the connection authorization
func (c *Conn) onConnect(packet *mqtt.Connect) bool {
	c.username = string(packet.Username)
	c.connect = &event.Connection{
		Peer:        c.service.ID(),
		Conn:        c.luid,
		WillFlag:    packet.WillFlag,
		WillRetain:  packet.WillRetainFlag,
		WillQoS:     packet.WillQOS,
		WillTopic:   packet.WillTopic,
		WillMessage: packet.WillMessage,
		ClientID:    packet.ClientID,
		Username:    packet.Username,
	}

	if c.service.cluster != nil {
		c.service.cluster.NotifyBeginOf(c.connect)
	}
	return true
}

// ------------------------------------------------------------------------------------

// onLink handles a request to create a link.
func (c *Conn) onLink(payload []byte) (response, bool) {
	var request linkRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		return errors.ErrBadRequest, false
	}

	// Check whether the name is a valid shortcut name
	if !shortcut.Match([]byte(request.Name)) {
		return errors.ErrLinkInvalid, false
	}

	// Ensures that the channel requested is valid
	channel := security.MakeChannel(request.Key, request.Channel)
	if channel == nil || channel.ChannelType == security.ChannelInvalid {
		return errors.ErrBadRequest, false
	}

	// Create the link with the name and set the full channel to it
	c.links[request.Name] = channel.String()

	// If an auto-subscribe was requested and the key has read permissions, subscribe
	if _, key, allowed := c.service.Authorize(channel, security.AllowRead); allowed && request.Subscribe {
		ssid := message.NewSsid(key.Contract(), channel.Query)
		c.service.pubsub.Subscribe(c, &event.Subscription{
			Conn:    c.LocalID(),
			User:    nocopy.String(c.Username()),
			Ssid:    ssid,
			Channel: channel.Channel,
		})
	}

	return &linkResponse{
		Status:  200,
		Name:    request.Name,
		Channel: channel.SafeString(),
	}, true
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

// -----------------------------------------------------------------------------------

// onKeyGen processes a keygen request.
func (c *Conn) onKeyGen(payload []byte) (response, bool) {
	message := keyGenRequest{}
	if err := json.Unmarshal(payload, &message); err != nil {
		return errors.ErrBadRequest, false
	}

	// Decrypt the parent key and make sure it's not expired
	parentKey, err := c.keys.DecryptKey(message.Key)
	if err != nil || parentKey.IsExpired() {
		return errors.ErrUnauthorized, false
	}

	// If the key provided is a master key, create a new key
	if parentKey.IsMaster() {
		key, err := c.keys.CreateKey(message.Key, message.Channel, message.access(), message.expires())
		if err != nil {
			return err, false
		}

		// Success, return the response
		return &keyGenResponse{
			Status:  200,
			Key:     key,
			Channel: message.Channel,
		}, true
	}

	// If the key provided can be extended, attempt to extend the key
	if parentKey.HasPermission(security.AllowExtend) {
		channel, err := c.keys.ExtendKey(message.Key, message.Channel, c.ID(), message.access(), message.expires())
		if err != nil {
			return err, false
		}

		// Success, return the response
		return &keyGenResponse{
			Status:  200,
			Key:     string(channel.Key),     // Encrypted channel key
			Channel: string(channel.Channel), // Channel name
		}, true
	}

	// Not authorised
	return errors.ErrUnauthorized, false
}

// -----------------------------------------------------------------------------------

// onKeyBan processes a keyban request.
func (c *Conn) onKeyBan(payload []byte) (response, bool) {
	message := keyBanRequest{}
	if err := json.Unmarshal(payload, &message); err != nil {
		return errors.ErrBadRequest, false
	}

	// Decrypt the secret key and make sure it's not expired and is a master key
	secretKey, err := c.keys.DecryptKey(message.Secret)
	if err != nil || secretKey.IsExpired() || !secretKey.IsMaster() {
		return errors.ErrUnauthorized, false
	}

	// Make sure the target key is for the same contract
	targetKey, err := c.keys.DecryptKey(message.Target)
	if err != nil || targetKey.Contract() != secretKey.Contract() {
		return errors.ErrBadRequest, false
	}

	// Depending on the flag, ban or unban the key
	bannedKey := event.Ban(message.Target)
	switch {
	case message.Banned && !c.service.cluster.Contains(&bannedKey):
		c.service.cluster.NotifyBeginOf(&bannedKey)
	case !message.Banned && c.service.cluster.Contains(&bannedKey):
		c.service.cluster.NotifyEndOf(&bannedKey)
	}

	// Success, return the response
	return &keyBanResponse{
		Status: 200,
		Banned: message.Banned,
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
				Username: conn.Username(),
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
	msg := presenceRequest{
		Status:  true, // Default: send status info
		Changes: nil,  // Default: send all changes
	}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return errors.ErrBadRequest, false
	}
	// Ensure we have trailing slash
	if !strings.HasSuffix(msg.Channel, "/") {
		msg.Channel = msg.Channel + "/"
	}
	// Parse the channel
	channel := security.ParseChannel([]byte(msg.Key + "/" + msg.Channel))
	if channel.ChannelType == security.ChannelInvalid {
		return errors.ErrBadRequest, false
	}
	// Check the authorization and permissions
	_, key, allowed := c.service.Authorize(channel, security.AllowPresence)
	if !allowed {
		return errors.ErrUnauthorized, false
	}

	// Keys which are supposed to be extended should not be used for presence
	if key.HasPermission(security.AllowExtend) {
		return errors.ErrUnauthorizedExt, false
	}

	// Create the ssid for the presence
	ssid := message.NewSsid(key.Contract(), channel.Query)

	// Check if the client is interested in subscribing/unsubscribing from changes.
	if msg.Changes != nil {
		ev := &event.Subscription{
			Conn:    c.LocalID(),
			User:    nocopy.String(c.Username()),
			Ssid:    message.NewSsidForPresence(ssid),
			Channel: channel.Channel,
		}

		switch *msg.Changes {
		case true:
			c.service.pubsub.Subscribe(c, ev)
		case false:
			c.service.pubsub.Unsubscribe(c, ev)
		}
	}

	// If we requested a status, populate the slice via scatter/gather.
	now := time.Now().UTC().Unix()
	who := make([]presenceInfo, 0, 4)
	if msg.Status {
		// Gather local & cluster presence
		who = append(who, getAllPresence(c.service, ssid)...)
		return &presenceResponse{
			Time:    now,
			Event:   presenceStatusEvent,
			Channel: msg.Channel,
			Who:     who,
		}, true
	}
	return nil, true
}
