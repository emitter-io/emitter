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

package event

import (
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/kelindar/binary"
	"github.com/kelindar/binary/nocopy"
)

// Various replicated event types.
const (
	typeSubscription = uint8(iota)
	typeBan
)

// Event represents an encodable event that happened at some point in time.
type Event interface {
	unitType() uint8
	Encode() string
}

// ------------------------------------------------------------------------------------

// Subscription represents a subscription event.
type Subscription struct {
	Peer    uint64        // The name of the peer. This must be first, since we're doing prefix search.
	Conn    security.ID   // The connection identifier.
	User    nocopy.String // The connection username.
	Channel nocopy.Bytes  // The channel string.
	Ssid    message.Ssid  // The SSID for the subscription.
}

// Type retuns the unit type.
func (e *Subscription) unitType() uint8 {
	return typeSubscription
}

// ConnID returns globally-unique identifier for the connection.
func (e *Subscription) ConnID() string {
	return e.Conn.Unique(uint64(e.Peer), "emitter")
}

// Encode encodes the event to string representation.
func (e *Subscription) Encode() string {
	buf, _ := binary.Marshal(e)
	return string(buf)
}

// decodeSubscription decodes the event
func decodeSubscription(encoded string) (Subscription, error) {
	var out Subscription
	return out, binary.Unmarshal([]byte(encoded), &out)
}

// ------------------------------------------------------------------------------------

// Ban represents a banned key event.
type Ban string

// Type retuns the unit type.
func (e *Ban) unitType() uint8 {
	return typeBan
}

// Encode encodes the event to string representation.
func (e Ban) Encode() string {
	return string(e)
}

// decodeBan decodes the event
func decodeBan(encoded string) (Ban, error) {
	return Ban(encoded), nil
}
