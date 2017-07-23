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

// Query represents an incoming query.
type Query struct {
	Name    string
	Respond func([]byte) error
}

// QueryResponse is used to represent a single response from a node
type QueryResponse struct {
	Node    string
	Payload []byte
}

// QueryEvent represents a generic query event sent by a node.
type QueryEvent struct {
	Node string // Gets or sets the node identifier for this event.
}

// SubscriptionEvent represents a message sent when a subscription is added or removed.
type SubscriptionEvent struct {
	Ssid    []uint32 // Gets or sets the SSID (parsed channel) for this subscription.
	Channel string   // Gets or sets the channel name.
	Node    string   // Gets or sets the node identifier for this event.
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

// decodeHandshakeEvent decodes the event from the decoder.
func decodeHandshakeEvent(decoder encoding.Decoder) (out *HandshakeEvent, err error) {
	out = new(HandshakeEvent)
	err = decoder.Decode(out)
	return
}

// MessageFrame represents a message frame which is sent through the wire to the
// remote server and contains a set of messages
type MessageFrame []*Message

// Message represents a message which has to be routed.
type Message struct {
	Ssid    []uint32 // The Ssid of the message
	Channel []byte   // The channel of the message
	Payload []byte   // The payload of the message
}

// decodeMessageFrame decodes the message frame from the decoder.
func decodeMessageFrame(decoder encoding.Decoder) (out MessageFrame, err error) {
	out = make(MessageFrame, 0, 64)
	err = decoder.Decode(&out)
	return
}
