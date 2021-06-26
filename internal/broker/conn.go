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

package broker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service/keygen"
	"github.com/emitter-io/stats"
	"github.com/kelindar/binary"
	"github.com/kelindar/binary/nocopy"
	"github.com/kelindar/rate"
)

const defaultReadRate = 100000

type response interface {
	ForRequest(uint16)
}

// Conn represents an incoming connection.
type Conn struct {
	sync.Mutex
	tracked  uint32            // Whether the connection was already tracked or not.
	socket   net.Conn          // The transport used to read and write messages.
	luid     security.ID       // The locally unique id of the connection.
	guid     string            // The globally unique id of the connection.
	service  *Service          // The service for this connection.
	subs     *message.Counters // The subscriptions for this connection.
	measurer stats.Measurer    // The measurer to use for monitoring.
	limit    *rate.Limiter     // The read rate limiter.
	keys     *keygen.Service   // The key generation provider.
	connect  *event.Connection // The associated connection event.
	username string            // The username provided by the client during MQTT connect.
	links    map[string]string // The map of all pre-authorized links.
}

// NewConn creates a new connection.
func (s *Service) newConn(t net.Conn, readRate int) *Conn {
	c := &Conn{
		tracked:  0,
		luid:     security.NewID(),
		service:  s,
		socket:   t,
		subs:     message.NewCounters(),
		measurer: s.measurer,
		links:    map[string]string{},
		keys:     s.keygen,
	}

	// Generate a globally unique id as well
	c.guid = c.luid.Unique(uint64(s.ID()), "emitter")
	if readRate == 0 {
		readRate = defaultReadRate
	}

	c.limit = rate.New(readRate, time.Second)

	// Increment the connection counter
	atomic.AddInt64(&s.connections, 1)
	return c
}

// ID returns the unique identifier of the subsriber.
func (c *Conn) ID() string {
	return c.guid
}

// LocalID returns the local connection identifier.
func (c *Conn) LocalID() security.ID {
	return c.luid
}

// Username returns the associated username.
func (c *Conn) Username() string {
	return c.username
}

// GetLink checks if the topic is a registered shortcut and expands it.
func (c *Conn) GetLink(topic []byte) []byte {
	if len(topic) <= 2 && c.links != nil {
		return []byte(c.links[binary.ToString(&topic)])
	}
	return topic
}

// AddLink adds a link alias for a channel.
func (c *Conn) AddLink(alias string, channel *security.Channel) {
	c.links[alias] = channel.String()
}

// Links returns a map of all links registered.
func (c *Conn) Links() map[string]string {
	return c.links
}

// Type returns the type of the subscriber
func (c *Conn) Type() message.SubscriberType {
	return message.SubscriberDirect
}

// MeasureElapsed measures elapsed time since
func (c *Conn) MeasureElapsed(name string, since time.Time) {
	c.measurer.MeasureElapsed(name, time.Now())
}

// Track tracks the connection by adding it to the metering.
func (c *Conn) Track(contract contract.Contract) {

	if atomic.CompareAndSwapUint32(&c.tracked, 0, 1) {
		// We keep only the IP address for fair tracking
		addr := c.socket.RemoteAddr().String()
		if tcp, ok := c.socket.RemoteAddr().(*net.TCPAddr); ok {
			addr = tcp.IP.String()
		}

		// Add the device to the stats and mark as done
		contract.Stats().AddDevice(addr)
	}
}

// Increment increments the subscription counter.
func (c *Conn) Increment(ssid message.Ssid, channel []byte) bool {
	return c.subs.Increment(ssid, channel)
}

// Decrement decrements a subscription counter.
func (c *Conn) Decrement(ssid message.Ssid) bool {
	return c.subs.Decrement(ssid)
}

// Process processes the messages.
func (c *Conn) Process() error {
	defer c.Close()
	reader := bufio.NewReaderSize(c.socket, 65536)
	maxSize := c.service.Config.MaxMessageBytes()
	for {
		// Set read/write deadlines so we can close dangling connections
		c.socket.SetDeadline(time.Now().Add(time.Second * 120))
		if c.limit.Limit() {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Decode an incoming MQTT packet
		msg, err := mqtt.DecodePacket(reader, maxSize)
		if err != nil {
			return err
		}

		// Handle the receive
		if err := c.onReceive(msg); err != nil {
			return err
		}
	}
}

// onReceive handles an MQTT receive.
func (c *Conn) onReceive(msg mqtt.Message) error {
	defer c.MeasureElapsed("rcv."+msg.String(), time.Now())
	switch msg.Type() {

	// We got an attempt to connect to MQTT.
	case mqtt.TypeOfConnect:
		var result uint8
		if !c.onConnect(msg.(*mqtt.Connect)) {
			result = 0x05 // Unauthorized
		}

		// Write the ack
		ack := mqtt.Connack{ReturnCode: result}
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
			if err := c.service.pubsub.OnSubscribe(c, sub.Topic); err != nil {
				ack.Qos = append(ack.Qos, 0x80) // 0x80 indicate subscription failure
				c.notifyError(err, packet.MessageID)
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
			if err := c.service.pubsub.OnUnsubscribe(c, sub.Topic); err != nil {
				c.notifyError(err, packet.MessageID)
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
		return io.EOF

	case mqtt.TypeOfPublish:
		packet := msg.(*mqtt.Publish)
		if err := c.service.pubsub.OnPublish(c, packet); err != nil {
			logging.LogError("conn", "publish received", err)
			c.notifyError(err, packet.MessageID)
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
		Header:  mqtt.Header{QOS: 0},
		Topic:   m.Channel, // The channel for this message.
		Payload: m.Payload, // The payload for this message.
	}

	_, err = packet.EncodeTo(c.socket)
	return
}

// notifyError notifies the connection about an error
func (c *Conn) notifyError(err *errors.Error, requestID uint16) {
	c.sendResponse("emitter/error/", err, requestID)
}

func (c *Conn) sendResponse(topic string, resp response, requestID uint16) {
	switch m := resp.(type) {
	case *errors.Error:
		cpy := m.Copy()
		cpy.ForRequest(requestID)
		resp = cpy
	default:
		m.ForRequest(requestID)
	}

	if b, err := json.Marshal(resp); err == nil {
		c.Send(&message.Message{
			Channel: []byte(topic),
			Payload: b,
		})
	}
	return
}

// CanSubscribe increments the internal counters and checks if the cluster
// needs to be notified.
func (c *Conn) CanSubscribe(ssid message.Ssid, channel []byte) bool {
	c.Lock()
	defer c.Unlock()
	return c.subs.Increment(ssid, channel)
}

// CanUnsubscribe decrements the internal counters and checks if the cluster
// needs to be notified.
func (c *Conn) CanUnsubscribe(ssid message.Ssid, channel []byte) bool {
	c.Lock()
	defer c.Unlock()
	return c.subs.Decrement(ssid)
}

// onConnect handles the connection authorization
func (c *Conn) onConnect(packet *mqtt.Connect) bool {
	c.username = string(packet.Username)
	c.connect = &event.Connection{
		Peer:        c.service.ID(),
		Conn:        c.luid,
		WillFlag:    packet.WillFlag,
		WillRetain:  packet.WillRetainFlag,
		WillQoS:     packet.WillQOS,
		WillTopic:   packet.WillTopic,
		WillMessage: packet.WillMessage,
		ClientID:    packet.ClientID,
		Username:    packet.Username,
	}

	if c.service.cluster != nil {
		c.service.cluster.Notify(c.connect, true)
	}
	return true
}

// Close terminates the connection.
func (c *Conn) Close() error {
	atomic.AddInt64(&c.service.connections, -1)
	if r := recover(); r != nil {
		logging.LogAction("closing", fmt.Sprintf("panic recovered: %s \n %s", r, debug.Stack()))
	}

	// Unsubscribe from everything, no need to lock since each Unsubscribe is
	// already locked. Locking the 'Close()' would result in a deadlock.
	for _, counter := range c.subs.All() {
		c.service.pubsub.Unsubscribe(c, &event.Subscription{
			Peer:    c.service.ID(),
			Conn:    c.luid,
			User:    nocopy.String(c.Username()),
			Ssid:    counter.Ssid,
			Channel: counter.Channel,
		})
	}

	// Publish last will
	c.service.pubsub.OnLastWill(c, c.connect)

	//logging.LogTarget("conn", "closed", c.guid)
	return c.socket.Close()
}
