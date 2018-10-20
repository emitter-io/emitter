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
	"encoding/json"
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/emitter-io/address"
	"github.com/emitter-io/emitter/message"
	"github.com/emitter-io/emitter/network/mqtt"
	"github.com/emitter-io/emitter/provider/contract"
	"github.com/emitter-io/emitter/provider/logging"
	"github.com/emitter-io/emitter/security"
	"github.com/emitter-io/stats"
)

// Conn represents an incoming connection.
type Conn struct {
	sync.Mutex
	tracked  uint32            // Whether the connection was already tracked or not.
	socket   net.Conn          // The transport used to read and write messages.
	username string            // The username provided by the client during MQTT connect.
	luid     security.ID       // The locally unique id of the connection.
	guid     string            // The globally unique id of the connection.
	service  *Service          // The service for this connection.
	subs     *message.Counters // The subscriptions for this connection.
	measurer stats.Measurer    // The measurer to use for monitoring.
}

// NewConn creates a new connection.
func (s *Service) newConn(t net.Conn) *Conn {
	c := &Conn{
		tracked:  0,
		luid:     security.NewID(),
		service:  s,
		socket:   t,
		subs:     message.NewCounters(),
		measurer: s.measurer,
	}

	// Generate a globally unique id as well
	c.guid = c.luid.Unique(uint64(address.GetHardware()), "emitter")
	logging.LogTarget("conn", "created", c.guid)

	// Increment the connection counter
	atomic.AddInt64(&s.connections, 1)
	return c
}

// ID returns the unique identifier of the subsriber.
func (c *Conn) ID() string {
	return c.guid
}

// Type returns the type of the subscriber
func (c *Conn) Type() message.SubscriberType {
	return message.SubscriberDirect
}

// MeasureElapsed measures elapsed time since
func (c *Conn) MeasureElapsed(name string, since time.Time) {
	c.measurer.MeasureElapsed(name, time.Now())
}

// track tracks the connection by adding it to the metering.
func (c *Conn) track(contract contract.Contract) {
	if atomic.LoadUint32(&c.tracked) == 0 {

		// We keep only the IP address for fair tracking
		addr := c.socket.RemoteAddr().String()
		if tcp, ok := c.socket.RemoteAddr().(*net.TCPAddr); ok {
			addr = tcp.IP.String()
		}

		// Add the device to the stats and mark as done
		contract.Stats().AddDevice(addr)
		atomic.StoreUint32(&c.tracked, 1)
	}
}

// Process processes the messages.
func (c *Conn) Process() error {
	defer c.Close()
	reader := bufio.NewReaderSize(c.socket, 65536)

	for {
		// Set read/write deadlines so we can close dangling connections
		c.socket.SetDeadline(time.Now().Add(time.Second * 120))

		// Decode an incoming MQTT packet
		msg, err := mqtt.DecodePacket(reader)
		if err != nil {
			return err
		}

		// Handle the receive
		if err := c.onReceive(msg); err != nil {
			return err
		}
	}
}

// notifyError notifies the connection about an error
func (c *Conn) notifyError(err *Error) {
	if b, err := json.Marshal(err); err == nil {
		c.Send(&message.Message{
			Channel: []byte("emitter/error/"),
			Payload: b,
		})
	}
}

// onReceive handles an MQTT receive.
func (c *Conn) onReceive(msg mqtt.Message) error {
	defer c.MeasureElapsed("rcv."+msg.String(), time.Now())
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
				ack.Qos = append(ack.Qos, 0x80) // 0x80 indicate subscription failure
				c.notifyError(err)
				continue
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
			if err := c.onUnsubscribe(sub.Topic); err != nil {
				c.notifyError(err)
			}
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
			logging.LogError("conn", "publish received", err)
			c.notifyError(err)
		}

		// Acknowledge the publication
		if packet.Header.QOS > 0 {
			ack := mqtt.Puback{MessageID: packet.MessageID}
			if _, err := ack.EncodeTo(c.socket); err != nil {
				return err
			}
		}
	}

	return nil
}

// Send forwards the message to the underlying client.
func (c *Conn) Send(m *message.Message) (err error) {
	defer c.MeasureElapsed("send.pub", time.Now())
	packet := mqtt.Publish{
		Header: &mqtt.StaticHeader{
			QOS: 0, // TODO when we'll support more QoS
		},
		MessageID: 0,         // TODO
		Topic:     m.Channel, // The channel for this message.
		Payload:   m.Payload, // The payload for this message.
	}

	// Acknowledge the publication
	_, err = packet.EncodeTo(c.socket)
	return
}

// Subscribe subscribes to a particular channel.
func (c *Conn) Subscribe(ssid message.Ssid, channel []byte) {
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
func (c *Conn) Unsubscribe(ssid message.Ssid, channel []byte) {
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
	if r := recover(); r != nil {
		logging.LogAction("closing", fmt.Sprintf("panic recovered: %s \n %s", r, debug.Stack()))
	}

	// Unsubscribe from everything, no need to lock since each Unsubscribe is
	// already locked. Locking the 'Close()' would result in a deadlock.
	for _, counter := range c.subs.All() {
		c.service.onUnsubscribe(counter.Ssid, c)
		c.service.notifyUnsubscribe(c, counter.Ssid, counter.Channel)
	}

	// Close the transport and decrement the connection counter
	atomic.AddInt64(&c.service.connections, -1)
	logging.LogTarget("conn", "closed", c.guid)
	return c.socket.Close()
}
