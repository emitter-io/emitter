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
	"time"

	"github.com/emitter-io/emitter/message"
	"github.com/emitter-io/emitter/provider/logging"
	"github.com/emitter-io/emitter/security"
)

// Error represents an event code which provides a more details.
type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// Error implements error interface.
func (e *Error) Error() string { return e.Message }

// Represents a set of errors used in the handlers.
var (
	ErrBadRequest      = &Error{Status: 400, Message: "The request was invalid or cannot be otherwise served."}
	ErrUnauthorized    = &Error{Status: 401, Message: "The security key provided is not authorized to perform this operation."}
	ErrPaymentRequired = &Error{Status: 402, Message: "The request can not be served, as the payment is required to proceed."}
	ErrForbidden       = &Error{Status: 403, Message: "The request is understood, but it has been refused or access is not allowed."}
	ErrNotFound        = &Error{Status: 404, Message: "The resource requested does not exist."}
	ErrServerError     = &Error{Status: 500, Message: "An unexpected condition was encountered and no more specific message is suitable."}
	ErrNotImplemented  = &Error{Status: 501, Message: "The server either does not recognize the request method, or it lacks the ability to fulfill the request."}
	ErrTargetInvalid   = &Error{Status: 400, Message: "Channel should end with `/` for strict types or `/#/` for wildcards."}
	ErrTargetTooLong   = &Error{Status: 400, Message: "Channel can not have more than 23 parts."}
)

// ------------------------------------------------------------------------------------

type keyGenRequest struct {
	Key     string `json:"key"`
	Channel string `json:"channel"`
	Type    string `json:"type"`
	TTL     int32  `json:"ttl"`
}

func (m *keyGenRequest) expires() time.Time {
	if m.TTL == 0 {
		return time.Unix(0, 0)
	}

	return time.Now().Add(time.Duration(m.TTL) * time.Second).UTC()
}

func (m *keyGenRequest) access() uint32 {
	required := security.AllowNone

	for i := 0; i < len(m.Type); i++ {
		switch c := m.Type[i]; c {
		case 'r':
			required |= security.AllowRead
		case 'w':
			required |= security.AllowWrite
		case 's':
			required |= security.AllowStore
		case 'l':
			required |= security.AllowLoad
		case 'p':
			required |= security.AllowPresence
		}
	}

	return required
}

// ------------------------------------------------------------------------------------

type meResponse struct {
	ID string `json:"id"`
}

// ------------------------------------------------------------------------------------

type keyGenResponse struct {
	Status  int    `json:"status"`
	Key     string `json:"key"`
	Channel string `json:"channel"`
}

// ------------------------------------------------------------------------------------

type presenceRequest struct {
	Key     string `json:"key"`     // The channel key for this request.
	Channel string `json:"channel"` // The target channel for this request.
	Status  bool   `json:"status"`  // Specifies that a status response should be sent.
	Changes bool   `json:"changes"` // Specifies that the changes should be notified.
}

type presenceEvent string

const (
	presenceStatusEvent      = presenceEvent("status")
	presenceSubscribeEvent   = presenceEvent("subscribe")
	presenceUnsubscribeEvent = presenceEvent("unsubscribe")
)

// ------------------------------------------------------------------------------------

// presenceNotify represents a state notification.
type presenceResponse struct {
	Time    int64          `json:"time"`    // The UNIX timestamp.
	Event   presenceEvent  `json:"event"`   // The event, must be "status", "subscribe" or "unsubscribe".
	Channel string         `json:"channel"` // The target channel for the notification.
	Who     []presenceInfo `json:"who"`     // The subscriber ids.
}

// ------------------------------------------------------------------------------------

// presenceInfo represents a presence info for a single connection.
type presenceInfo struct {
	ID       string `json:"id"`                 // The subscriber ID.
	Username string `json:"username,omitempty"` // The subscriber username set by client ID.
}

// ------------------------------------------------------------------------------------

// presenceNotify represents a state notification.
type presenceNotify struct {
	Time    int64         `json:"time"`    // The UNIX timestamp.
	Event   presenceEvent `json:"event"`   // The event, must be "status", "subscribe" or "unsubscribe".
	Channel string        `json:"channel"` // The target channel for the notification.
	Who     presenceInfo  `json:"who"`     // The subscriber id.
	Ssid    message.Ssid  `json:"-"`       // The ssid to dispatch the notification on.
}

// newPresenceNotify creates a new notification payload.
func newPresenceNotify(ssid message.Ssid, event presenceEvent, channel string, id string, username string) *presenceNotify {
	return &presenceNotify{
		Ssid:    message.NewSsidForPresence(ssid),
		Time:    time.Now().UTC().Unix(),
		Event:   event,
		Channel: channel,
		Who: presenceInfo{
			ID:       id,
			Username: username,
		},
	}
}

// Encode encodes the presence notifications and returns a payload to send.
func (e *presenceNotify) Encode() ([]byte, bool) {
	encoded, err := json.Marshal(e)
	if err != nil {
		logging.LogError("presence", "encoding presence notification", err)
		return nil, false
	}

	return encoded, true
}
