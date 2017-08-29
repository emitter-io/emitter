package broker

import (
	"encoding/json"
	"time"

	"github.com/emitter-io/emitter/security"
)

// EventError represents an event code which provides a more de.
type EventError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// Error implements error interface.
func (e *EventError) Error() string { return e.Message }

// Represents a set of errors used in the handlers.
var (
	ErrBadRequest      = &EventError{Status: 400, Message: "The request was invalid or cannot be otherwise served."}
	ErrUnauthorized    = &EventError{Status: 401, Message: "The security key provided is not authorized to perform this operation."}
	ErrPaymentRequired = &EventError{Status: 402, Message: "The request can not be served, as the payment is required to proceed."}
	ErrForbidden       = &EventError{Status: 403, Message: "The request is understood, but it has been refused or access is not allowed."}
	ErrNotFound        = &EventError{Status: 404, Message: "The resource requested does not exist."}
	ErrServerError     = &EventError{Status: 500, Message: "An unexpected condition was encountered and no more specific message is suitable."}
	ErrNotImplemented  = &EventError{Status: 501, Message: "The server either does not recognize the request method, or it lacks the ability to fulfill the request."}
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

	return time.Now().Add(time.Second).UTC()
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

type keyGenResponse struct {
	Status  int    `json:"status"`
	Key     string `json:"key"`
	Channel string `json:"channel"`
}

// ------------------------------------------------------------------------------------

type presenceEvent string

const (
	presenceStatusEvent      = presenceEvent("status")
	presenceSubscribeEvent   = presenceEvent("subscribe")
	presenceUnsubscribeEvent = presenceEvent("unsubscribe")
)

// presenceNotify represents a state notification.
type presenceNotify struct {
	Time    int64         `json:"time"`    // The UNIX timestamp.
	Event   presenceEvent `json:"event"`   // The event, must be "status", "subscribe" or "unsubscribe".
	Channel string        `json:"channel"` // The target channel for the notification.
	Who     string        `json:"who"`     // The subscriber id.
}

// newPresenceNotify creates a new notification payload.
func newPresenceNotify(event presenceEvent, channel string, who string) (msg []byte) {
	msg, _ = json.Marshal(presenceNotify{
		Time:    time.Now().UTC().Unix(),
		Event:   event,
		Channel: channel,
		Who:     who,
	})
	return
}
