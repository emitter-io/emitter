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
	Key string `json:"key"` // The channel key for this request.
	// TODO: Target instead of Channel
	Channel       string     `json:"channel"` // The target channel for this request.
	LastMessageID message.ID `json:"lastMessageID,omitempty"`
}

type Message struct {
	ID      message.ID `json:"id"`
	Topic   string     `json:"topic"`   // The channel of the message
	Payload string     `json:"payload"` // The payload of the message
}
type Response struct {
	Request  uint16    `json:"req,omitempty"` // The corresponding request ID.
	Messages []Message `json:"messages"`      // The history of messages.
	//Messages message.Frame `json:"messages"`
}

// ForRequest sets the request ID in the response for matching
func (r *Response) ForRequest(id uint16) {
	r.Request = id
}

type limiter struct {
	maxCount     int64
	currentCount int64
	totalSize    int64
}

func (l *limiter) CanAddMessage(m message.Message) bool {
	if l.currentCount >= l.maxCount {
		return false
	}
	return true
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

	limit := int64(3)
	if v, ok := channel.Last(); ok {
		limit = v
	}
	messageLimiter := &limiter{
		maxCount: limit,
	}

	ssid := message.NewSsid(key.Contract(), channel.Query)
	t0, t1 := channel.Window() // Get the window

	msgs, err := s.store.LimitedQuery(ssid, t0, t1, messageLimiter, request.LastMessageID)
	if err != nil {
		logging.LogError("conn", "query last messages", err)
		return errors.ErrServerError, false
	}

	// This request is answered either by resending all messages on their
	// original channel, potentially triggering mutliple handlers on the client
	// side, or by responding with all messages in one big response here.
	// Can be both, but the latter is the default behavior.
	withResponse, okResponse := channel.GetOption("response")
	withResend, okResend := channel.GetOption("resend")
	doResend := okResend && withResend == 1
	doRespond := (okResponse && withResponse == 1) || !doResend

	// Resend every messages again like they were originally.
	/*
		if doResend {
			// Range over the messages in the channel and forward them
			for _, m := range msgs {
				msg := m // Copy message
				c.Send(&msg)
			}
		}*/

	// Send all messages in the payload of the response to this request.
	if doRespond {
		resp := &Response{
			Messages: make([]Message, 0, len(msgs)),
		}
		for _, m := range msgs {
			msg := m
			resp.Messages = append(resp.Messages, Message{
				ID:      msg.ID,
				Topic:   string(msg.Channel), // The channel for this message.
				Payload: string(msg.Payload), // The payload for this message.
			})
		}
		return resp, true
	}

	return nil, true
}
