package broker

import (
	"time"

	"github.com/emitter-io/emitter/security"
)

// EventError represents an event code which provides a more de.
type EventError struct {
	code int
	msg  string
}

// Error implements error interface.
func (e *EventError) Error() string { return e.msg }

// Represents a set of errors used in the handlers.
var (
	ErrBadRequest      = &EventError{code: 400, msg: "The request was invalid or cannot be otherwise served."}
	ErrUnauthorized    = &EventError{code: 401, msg: "The security key provided is not authorized to perform this operation."}
	ErrPaymentRequired = &EventError{code: 402, msg: "The request can not be served, as the payment is required to proceed."}
	ErrForbidden       = &EventError{code: 403, msg: "The request is understood, but it has been refused or access is not allowed."}
	ErrNotFound        = &EventError{code: 404, msg: "The resource requested does not exist."}
	ErrServerError     = &EventError{code: 500, msg: "An unexpected condition was encountered and no more specific message is suitable."}
	ErrNotImplemented  = &EventError{code: 501, msg: "The server either does not recognize the request method, or it lacks the ability to fulfill the request."}
)

// ------------------------------------------------------------------------------------

type keyGenMessage struct {
	Key     string `json:"key"`
	Channel string `json:"channel"`
	Type    string `json:"type"`
	TTL     int32  `json:"ttl"`
}

func (m *keyGenMessage) expires() time.Time {
	if m.TTL == 0 {
		return time.Unix(0, 0)
	}

	return time.Now().Add(time.Second).UTC()
}

func (m *keyGenMessage) access() uint32 {
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
