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

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/security"
)

// Response represents a response which can be sent for a specific request.
type response interface {
	ForRequest(uint16)
}

// ------------------------------------------------------------------------------------

// Error represents an event code which provides a more details.
type Error struct {
	Request uint16 `json:"req,omitempty"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// Error implements error interface.
func (e *Error) Error() string { return e.Message }

// ForRequest sets the request ID in the response for matching
func (e *Error) ForRequest(id uint16) {
	e.Request = id
}

// Represents a set of errors used in the handlers.
var (
	ErrBadRequest      = &Error{Status: 400, Message: "the request was invalid or cannot be otherwise served"}
	ErrUnauthorized    = &Error{Status: 401, Message: "the security key provided is not authorized to perform this operation"}
	ErrPaymentRequired = &Error{Status: 402, Message: "the request can not be served, as the payment is required to proceed"}
	ErrForbidden       = &Error{Status: 403, Message: "the request is understood, but it has been refused or access is not allowed"}
	ErrNotFound        = &Error{Status: 404, Message: "the resource requested does not exist"}
	ErrServerError     = &Error{Status: 500, Message: "an unexpected condition was encountered and no more specific message is suitable"}
	ErrNotImplemented  = &Error{Status: 501, Message: "the server either does not recognize the request method, or it lacks the ability to fulfill the request"}
	ErrTargetInvalid   = &Error{Status: 400, Message: "channel should end with `/` for strict types or `/#/` for wildcards"}
	ErrTargetTooLong   = &Error{Status: 400, Message: "channel can not have more than 23 parts"}
	ErrLinkInvalid     = &Error{Status: 400, Message: "the link must be an alphanumeric string of 1 or 2 characters"}
)

// ------------------------------------------------------------------------------------

type keyGenRequest struct {
	Key     string `json:"key"`     // The master key to use.
	Channel string `json:"channel"` // The channel to create a key for.
	Type    string `json:"type"`    // The permission set.
	TTL     int32  `json:"ttl"`     // The TTL of the key.
}

func (m *keyGenRequest) expires() time.Time {
	if m.TTL == 0 {
		return time.Unix(0, 0)
	}

	return time.Now().Add(time.Duration(m.TTL) * time.Second).UTC()
}

func (m *keyGenRequest) access() uint8 {
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
		case 'e':
			required |= security.AllowExtend
		case 'x':
			required |= security.AllowExecute
		}
	}

	return required
}

// ------------------------------------------------------------------------------------

type keyGenResponse struct {
	Request uint16 `json:"req,omitempty"`
	Status  int    `json:"status"`
	Key     string `json:"key"`
	Channel string `json:"channel"`
}

// ForRequest sets the request ID in the response for matching
func (r *keyGenResponse) ForRequest(id uint16) {
	r.Request = id
}

// ------------------------------------------------------------------------------------

type linkRequest struct {
	Name      string `json:"name"`      // The name of the shortcut, max 2 characters.
	Key       string `json:"key"`       // The key for the channel.
	Channel   string `json:"channel"`   // The channel name for the shortcut.
	Subscribe bool   `json:"subscribe"` // Specifies whether the broker should auto-subscribe.
	Private   bool   `json:"private"`   // Specifies whether the broker should generate a private link.
}

// ------------------------------------------------------------------------------------

type linkResponse struct {
	Request uint16 `json:"req,omitempty"`     // The corresponding request ID.
	Status  int    `json:"status"`            // The status of the response.
	Name    string `json:"name,omitempty"`    // The name of the shortcut, max 2 characters.
	Channel string `json:"channel,omitempty"` // The channel which was registered.
}

// ForRequest sets the request ID in the response for matching
func (r *linkResponse) ForRequest(id uint16) {
	r.Request = id
}

// ------------------------------------------------------------------------------------

type meResponse struct {
	Request uint16            `json:"req,omitempty"`   // The corresponding request ID.
	ID      string            `json:"id"`              // The private ID of the connection.
	Links   map[string]string `json:"links,omitempty"` // The set of pre-defined channels.
}

// ForRequest sets the request ID in the response for matching
func (r *meResponse) ForRequest(id uint16) {
	r.Request = id
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
	Request uint16         `json:"req,omitempty"` // The corresponding request ID.
	Time    int64          `json:"time"`          // The UNIX timestamp.
	Event   presenceEvent  `json:"event"`         // The event, must be "status", "subscribe" or "unsubscribe".
	Channel string         `json:"channel"`       // The target channel for the notification.
	Who     []presenceInfo `json:"who"`           // The subscriber ids.
}

// ForRequest sets the request ID in the response for matching
func (r *presenceResponse) ForRequest(id uint16) {
	r.Request = id
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
