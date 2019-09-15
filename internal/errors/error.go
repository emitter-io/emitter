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

package errors

// New creates a new error
func New(msg string) *Error {
	return &Error{
		Status:  500,
		Message: msg,
	}
}

// Error represents an event code which provides a more details.
type Error struct {
	Request uint16 `json:"req,omitempty"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// Error implements error interface.
func (e *Error) Error() string { return e.Message }

// Copy clones the error object.
func (e *Error) Copy() *Error {
	copyErr := *e
	return &copyErr
}

// ForRequest returns an error for a specific request.
func (e *Error) ForRequest(requestID uint16) {
	e.Request = requestID
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
	ErrUnauthorizedExt = &Error{Status: 401, Message: "the security key with extend permission can only be used for private links"}
)
