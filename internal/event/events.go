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
	typeSub = uint8(iota)
	typeBan
	typeConn
)

// Event represents an encodable event that happened at some point in time.
type Event interface {
	unitType() uint8
	Key() string
	Val() []byte
}

// ------------------------------------------------------------------------------------

// Subscription represents a subscription event.
type Subscription struct {
	Peer    uint64        `binary:"-"` // The name of the peer. This must be first, since we're doing prefix search.
	Conn    security.ID   `binary:"-"` // The connection identifier.
	Ssid    message.Ssid  `binary:"-"` // The SSID for the subscription.
	User    nocopy.String // The connection username.
	Channel nocopy.Bytes  // The channel string.
}

// Type retuns the unit type.
func (e *Subscription) unitType() uint8 {
	return typeSub
}

// ConnID returns globally-unique identifier for the connection.
func (e *Subscription) ConnID() string {
	return e.Conn.Unique(uint64(e.Peer), "emitter")
}

// Key returns the event key.
func (e *Subscription) Key() string {
	buffer := make([]byte, 16+4*len(e.Ssid))
	binary.BigEndian.PutUint64(buffer[0:8], e.Peer)
	binary.BigEndian.PutUint64(buffer[8:16], uint64(e.Conn))
	for i := 0; i < len(e.Ssid); i++ {
		binary.BigEndian.PutUint32(buffer[16+(i*4):20+(i*4)], e.Ssid[i])
	}
	return binary.ToString(&buffer)
}

// Val returns the event value.
func (e *Subscription) Val() []byte {
	buffer, _ := binary.Marshal(e)
	return buffer
}

// decodeSubscription decodes the event
func decodeSubscription(k string, v []byte) (e Subscription, err error) {
	if len(v) > 0 {
		err = binary.Unmarshal(v, &e)
	}

	// Decode the key
	buffer := binary.ToBytes(k)
	e.Peer = binary.BigEndian.Uint64(buffer[0:8])
	e.Conn = security.ID(binary.BigEndian.Uint64(buffer[8:16]))
	e.Ssid = make(message.Ssid, (len(buffer)-16)/4)
	for i := 0; i < len(e.Ssid); i++ {
		e.Ssid[i] = binary.BigEndian.Uint32(buffer[16+(i*4) : 20+(i*4)])
	}

	return e, err
}

// ------------------------------------------------------------------------------------

// Ban represents a banned key event.
type Ban string

// Type retuns the unit type.
func (e *Ban) unitType() uint8 {
	return typeBan
}

// Key returns the event key.
func (e Ban) Key() string {
	return string(e)
}

// Val returns the event value.
func (e Ban) Val() []byte {
	return nil
}

// decodeBan decodes the event
func decodeBan(k string) (Ban, error) {
	return Ban(k), nil
}

// ------------------------------------------------------------------------------------

// Connection represents a banned key event.
type Connection struct {
	Peer        uint64      `binary:"-"` // The name of the peer. This must be first, since we're doing prefix search.
	Conn        security.ID `binary:"-"` // The connection identifier.
	WillFlag    bool
	WillRetain  bool
	WillQoS     uint8
	WillTopic   []byte
	WillMessage []byte
	ClientID    []byte
	Username    []byte
}

// Type retuns the unit type.
func (e *Connection) unitType() uint8 {
	return typeConn
}

// Key returns the event key.
func (e Connection) Key() string {
	buffer := make([]byte, 16)
	binary.BigEndian.PutUint64(buffer[0:8], e.Peer)
	binary.BigEndian.PutUint64(buffer[8:16], uint64(e.Conn))
	return binary.ToString(&buffer)
}

// Val returns the event value.
func (e Connection) Val() []byte {
	buffer, _ := binary.Marshal(e)
	return buffer
}

// decodeConnection decodes the event
func decodeConnection(k string, v []byte) (e Connection, err error) {
	if len(v) > 0 {
		err = binary.Unmarshal(v, &e)
	}

	// Decode the key
	buffer := binary.ToBytes(k)
	e.Peer = binary.BigEndian.Uint64(buffer[0:8])
	e.Conn = security.ID(binary.BigEndian.Uint64(buffer[8:16]))
	return e, err
}
