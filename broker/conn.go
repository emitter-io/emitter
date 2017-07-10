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
	"net"
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/mqtt"
	"github.com/emitter-io/emitter/security"
)

var logConnection = logging.AddLogger("[conn] connection %s id=%u64")

// nextIdentifier generates next identifier and uses the clock to initialize
// the counter for reboot tolerance.
var nextIdentifier = func() func() uint64 {
	id := uint64(time.Now().UTC().Unix())
	return func() uint64 {
		return atomic.AddUint64(&id, 1)
	}
}()

// Conn represents an incoming connection.
type Conn struct {
	socket  net.Conn                 // The transport used to read and write messages.
	id      uint64                   // The identifier of the connection.
	service *Service                 // The service for this connection.
	subs    map[uint32]*Subscription // The subscriptions for this connection.
}

// NewConn creates a new connection.
func (s *Service) newConn(t net.Conn) *Conn {
	c := &Conn{
		id:      nextIdentifier(),
		service: s,
		socket:  t,
	}

	logging.Log(logConnection, "created", c.id)
	return c
}

// Process processes the messages.
func (c *Conn) Process() {
	defer c.Close()

	for {
		msg, err := mqtt.DecodePacket(c.socket)
		if err != nil {
			logging.LogError("conn", "read from transport", err)
			break
		}

		println(msg.Type())

		switch msg.Type() {

		// We got an attempt to connect to MQTT.
		case mqtt.TypeOfConnect:
			//packet := msg.(*mqtt.Connect)

			// Write the ack
			ack := mqtt.Connack{ReturnCode: 0x00}
			if _, err := ack.EncodeTo(c.socket); err != nil {
				logging.LogError("conn", "connack send", err)
				break
			}

		// We got an attempt to subscribe to a channel.
		case mqtt.TypeOfSubscribe:
			packet := msg.(*mqtt.Subscribe)
			ack := mqtt.Suback{
				MessageID: packet.MessageID,
				Qos:       make([]uint8, 0, len(packet.Subscriptions)),
			}

			// Subscribe for each subscription
			for _, sub := range packet.Subscriptions {
				if err := c.onSubscribe(sub.Topic); err != nil {
					// TODO: Handle Error
				}

				// Append the QoS
				ack.Qos = append(ack.Qos, sub.Qos)
			}

			// Acknowledge the subscription
			if _, err := ack.EncodeTo(c.socket); err != nil {
				logging.LogError("conn", "suback send", err)
				break
			}

		// We got an attempt to unsubscribe from a channel.
		case mqtt.TypeOfUnsubscribe:
			packet := msg.(*mqtt.Unsubscribe)
			ack := mqtt.Unsuback{MessageID: packet.MessageID}

			// Unsubscribe from each subscription
			for _, sub := range packet.Topics {
				c.onUnsubscribe(sub.Topic) // TODO: Handle error or just ignore?
			}

			// Acknowledge the unsubscription
			if _, err := ack.EncodeTo(c.socket); err != nil {
				logging.LogError("conn", "unsuback send", err)
				break
			}

		// We got an MQTT ping response, respond appropriately.
		case mqtt.TypeOfPingreq:
			ack := mqtt.Pingresp{}
			if _, err := ack.EncodeTo(c.socket); err != nil {
				logging.LogError("conn", "connack send", err)
				break
			}

		case mqtt.TypeOfDisconnect:

		case mqtt.TypeOfPublish:
			packet := msg.(*mqtt.Publish)
			/*ack := mqtt.Puback{
				MessageID: packet.MessageID,
			}*/

			if err := c.onPublish(packet.Topic, packet.Payload); err != nil {
				// TODO: Handle Error
			}

			// Acknowledge the publication
			/*if _, err := ack.EncodeTo(c.socket); err != nil {
				logging.LogError("conn", "puback send", err)
				break
			}*/

		}
	}
}

// Send forwards the message to the underlying client.
func (c *Conn) Send(channel []byte, payload []byte) error {
	packet := mqtt.Publish{
		Header: &mqtt.StaticHeader{
			QOS: 0, // TODO when we'll support more QoS
		},
		MessageID: 0,       // TODO
		Topic:     channel, // The channel for this message.
		Payload:   payload, // The payload for this message.
	}

	println("sending message to " + string(channel))

	// Acknowledge the publication
	if _, err := packet.EncodeTo(c.socket); err != nil {
		logging.LogError("conn", "message send", err)
		return err
	}

	return nil
}

// Subscribe subscribes to a particular channel.
func (c *Conn) Subscribe(contract int32, channel *security.Channel) {
	ssid := NewSsid(contract, channel)
	hkey := ssid.GetHashCode()

	// Only subscribe if we don't yet have a subscription
	if _, exists := c.subs[hkey]; exists {
		return
	}

	// Add the subscription
	if sub, err := c.service.subscriptions.Subscribe(ssid, string(channel.Channel), c); err != nil {
		c.subs[hkey] = sub

		// TODO: Broadcast the subscription within our cluster
		//c.service.Broadcast("SUB", []byte(topic), true) // TODO: also include node id, probably can just use msgpack we've got.
	}
}

// Unsubscribe unsubscribes this client from a particular channel.
func (c *Conn) Unsubscribe(ssid Ssid) {
	hkey := ssid.GetHashCode()

	// Get the subscription from our internal map
	if sub, ok := c.subs[hkey]; ok {
		// Unsubscribe from the trie and remove from our internal map
		c.service.subscriptions.Unsubscribe(sub)
		delete(c.subs, hkey)

		// TODO: Broadcast the unsubscription within our cluster
		//c.service.Broadcast("UNSUB", []byte(channel.Channel), true) // TODO: also include node id, probably can just use msgpack we've got.
	}
}

// Close terminates the connection.
func (c *Conn) Close() error {
	logging.Log(logConnection, "closed", c.id)

	// Unsubscribe from everything. TODO: Lock this?
	for _, s := range c.subs {
		c.Unsubscribe(s.Ssid)
	}

	return c.socket.Close()
}
