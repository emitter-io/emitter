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

package history

import (
	"encoding/json"

	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service"
)

// Request represents a historical messages request.
type Request struct {
	Key     string `json:"key"`     // The channel key for this request.
	Channel string `json:"channel"` // The target channel for this request.
}

// OnRequest handles a request of historical messages.
func (s *Service) OnRequest(c service.Conn, payload []byte) (service.Response, bool) {
	var request Request
	if err := json.Unmarshal(payload, &request); err != nil {
		return errors.ErrBadRequest, false
	}

	channel := security.ParseChannel([]byte(request.Channel))
	if channel.ChannelType == security.ChannelInvalid {
		return errors.ErrBadRequest, false
	}

	// Check the authorization and permissions
	_, key, allowed := s.auth.Authorize(channel, security.AllowLoad)
	if !allowed {
		return errors.ErrUnauthorized, false
	}

	limit := int64(0)
	if v, ok := channel.Last(); ok {
		limit = v
	}

	ssid := message.NewSsid(key.Contract(), channel.Query)
	t0, t1 := channel.Window() // Get the window
	msgs, err := s.store.Query(ssid, t0, t1, int(limit))
	if err != nil {
		logging.LogError("conn", "query last messages", err)
		return errors.ErrServerError, false
	}

	// Range over the messages in the channel and forward them
	for _, m := range msgs {
		msg := m // Copy message
		c.Send(&msg)
	}

	return nil, true
}
