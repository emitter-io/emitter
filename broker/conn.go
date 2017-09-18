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
	"bufio"
	"fmt"
	"net"
	"runtime/debug"
	"sync"

	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/emitter/network/mqtt"
	"github.com/emitter-io/emitter/security"
)

// Conn represents an incoming connection.
type Conn struct {
	sync.Mutex
	socket   net.Conn               // The transport used to read and write messages.
	username string                 // The username provided by the client during MQTT connect.
	luid     security.ID            // The locally unique id of the connection.
	guid     string                 // The globally unique id of the connection.
	service  *Service               // The service for this connection.
	subs     *subscription.Counters // The subscriptions for this connection.
}

// NewConn creates a new connection.
func (s *Service) newConn(t net.Conn) *Conn {
	c := &Conn{
		luid:    security.NewID(),
		service: s,
		socket:  t,
		subs:    subscription.NewCounters(),
	}

	// Generate a globally unique id as well
	c.guid = c.luid.Unique(uint64(address.Hardware()), "emitter")
	logging.LogTarget("conn", "created", c.luid)
	return c
}

// ID returns the unique identifier of the subsriber.
func (c *Conn) ID() string {
	return c.guid
}

// Type returns the type of the subscriber
func (c *Conn) Type() subscription.SubscriberType {
	return subscription.SubscriberDirect
}

// Process processes the messages.
func (c *Conn) Process() error {
	defer c.Close()
	reader := bufio.NewReaderSize(c.socket, 65536)

	for {
		// Decode an incoming MQTT packet
		msg, err := mqtt.DecodePacket(reader)
		if err != nil {
			return err
		}

		switch msg.Type() {

		// We got an attempt to connect to MQTT.
		case mqtt.TypeOfConnect:
			packet := msg.(*mqtt.Connect)
			c.username = string(packet.Username)

			// Write the ack
			ack := mqtt.Connack{ReturnCode: 0x00}
			if _, err := ack.EncodeTo(c.socket); err != nil {
				return err
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
				return err
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
				return err
			}

		// We got an MQTT ping response, respond appropriately.
		case mqtt.TypeOfPingreq:
			ack := mqtt.Pingresp{}
			if _, err := ack.EncodeTo(c.socket); err != nil {
				return err
			}

		case mqtt.TypeOfDisconnect:
			return nil

		case mqtt.TypeOfPublish:
			packet := msg.(*mqtt.Publish)

			if err := c.onPublish(packet.Topic, packet.Payload); err != nil {
				// TODO: Handle Error
				println(err.Error())
			}

			// Acknowledge the publication
			if packet.Header.QOS > 0 {
				ack := mqtt.Puback{MessageID: packet.MessageID}
				if _, err := ack.EncodeTo(c.socket); err != nil {
					return err
				}
			}
		}
	}
}

// Send forwards the message to the underlying client.
func (c *Conn) Send(ssid subscription.Ssid, channel []byte, payload []byte) error {
	packet := mqtt.Publish{
		Header: &mqtt.StaticHeader{
			QOS: 0, // TODO when we'll support more QoS
		},
		MessageID: 0,       // TODO
		Topic:     channel, // The channel for this message.
		Payload:   payload, // The payload for this message.
	}

	// Acknowledge the publication
	_, err := packet.EncodeTo(c.socket)
	if err != nil {
		logging.LogError("conn", "message send", err)
		return err
	}

	return nil
}

// Subscribe subscribes to a particular channel.
func (c *Conn) Subscribe(ssid subscription.Ssid, channel []byte) {
	c.Lock()
	defer c.Unlock()

	// Add the subscription
	if first := c.subs.Increment(ssid, channel); first {

		// Subscribe the subscriber
		c.service.onSubscribe(ssid, c)

		// Broadcast the subscription within our cluster
		c.service.notifySubscribe(c, ssid, channel)
	}
}

// Unsubscribe unsubscribes this client from a particular channel.
func (c *Conn) Unsubscribe(ssid subscription.Ssid, channel []byte) {
	c.Lock()
	defer c.Unlock()

	// Decrement the counter and if there's no more subscriptions, notify everyone.
	if last := c.subs.Decrement(ssid); last {

		// Unsubscribe the subscriber
		c.service.onUnsubscribe(ssid, c)

		// Broadcast the unsubscription within our cluster
		c.service.notifyUnsubscribe(c, ssid, channel)
	}
}

// Close terminates the connection.
func (c *Conn) Close() error {
	logging.LogTarget("conn", "closed", c.luid)

	// Unsubscribe from everything, no need to lock since each Unsubscribe is
	// already locked. Locking the 'Close()' would result in a deadlock.
	for _, counter := range c.subs.All() {
		c.service.onUnsubscribe(counter.Ssid, c)
		c.service.notifyUnsubscribe(c, counter.Ssid, counter.Channel)
	}

	// Attempt to recover a panic
	if r := recover(); r != nil {
		logging.LogAction("closing", fmt.Sprintf("pancic recovered: %s \n %s", r, debug.Stack()))
	}

	// Close the transport
	return c.socket.Close()
}
