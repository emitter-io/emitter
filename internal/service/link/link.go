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

package link

import (
	"encoding/json"
	"regexp"

	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service"
	"github.com/kelindar/binary/nocopy"
)

var (
	shortcut = regexp.MustCompile("^[a-zA-Z0-9]{1,2}$")
)

// Service represents a link generation service.
type Service struct {
	auth   service.Authorizer // The authorizer to use.
	pubsub service.PubSub     // The pub/sub service to use.
}

// New creates a new link generation service.
func New(auth service.Authorizer, pubsub service.PubSub) *Service {
	return &Service{
		auth:   auth,
		pubsub: pubsub,
	}
}

// OnRequest handles a request to create a link.
func (s *Service) OnRequest(c service.Conn, payload []byte) (service.Response, bool) {
	var request Request
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
	c.AddLink(request.Name, channel)

	// If an auto-subscribe was requested and the key has read permissions, subscribe
	if _, key, allowed := s.auth.Authorize(channel, security.AllowRead); allowed && request.Subscribe {
		ssid := message.NewSsid(key.Contract(), channel.Query)
		s.pubsub.Subscribe(c, &event.Subscription{
			Conn:    c.LocalID(),
			User:    nocopy.String(c.Username()),
			Ssid:    ssid,
			Channel: channel.Channel,
		})
	}

	return &Response{
		Status:  200,
		Name:    request.Name,
		Channel: channel.SafeString(),
	}, true
}
