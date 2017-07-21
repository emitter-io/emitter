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

package cluster

import (
	"github.com/emitter-io/emitter/encoding"
)

// SubscriptionEvent represents a message sent when a subscription is added or removed.
type SubscriptionEvent struct {
	Ssid []uint32 // Gets or sets the SSID (parsed channel) for this subscription.
	Node string   // Gets or sets the node identifier for this event.
}

// decodeSubscriptionEvent decodes the event from the payload.
func decodeSubscriptionEvent(payload []byte) *SubscriptionEvent {
	var event SubscriptionEvent
	encoding.Decode(payload, &event)
	return &event
}

// HandshakeEvent represents a message used to confirm the identity of a remote server
// which is trying to connect.
type HandshakeEvent struct {
	Key  string // Gets or sets the handshake key to verify.
	Node string // Gets or sets the node identifier for this event.
}

// decodeHandshakeEvent decodes the event from the payload.
func decodeHandshakeEvent(payload []byte) *HandshakeEvent {
	var event HandshakeEvent
	encoding.Decode(payload, &event)
	return &event
}
