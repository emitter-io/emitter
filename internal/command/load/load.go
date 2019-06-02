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

package load

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/jawher/mow.cli"
)

const maxMessageSize = 64000

var dial = net.Dial

// Run runs a benchmark command.
func Run(cmd *cli.Cmd) {
	cmd.Spec = "KEY [ -h=<host> ] [ -c=<channel> ] [ -b=<batch> ] [ -s=<size> ]"
	var (
		key     = cmd.StringArg("KEY", "", "Key for the channel (by default a key for the `load/` channel)")
		host    = cmd.StringOpt("h host", "127.0.0.1:8080", "Host name")
		channel = cmd.StringOpt("c channel", "load/", "The channel for load testing")
		batch   = cmd.IntOpt("b batch", 10, "The size of the batch to send")
		size    = cmd.IntOpt("s size", 64, "The size of the message to send")
	)
	cmd.Action = func() {
		logging.LogAction("client", "starting a load test...")
		cli, err := newConn(*host, *key, *channel)
		if err != nil {
			logging.LogError("client", "connection to the broker", err)
			return
		}

		defer cli.Close()
		go cli.Drain()
		logging.LogAction("client", "draining messages...")
		msg := newMessage(cli.topic, *size)
		for {
			for i := 0; i < *batch; i++ {
				if _, err := msg.EncodeTo(cli); err != nil {
					logging.LogError("client", "tcp send", err)
					return
				}
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// newMessage creates a new MQTT message for the load test.
func newMessage(topic string, size int) mqtt.Publish {
	// "4kzJv3TMhYTg6lLk6fQoFG2KCe7gjFPk/a/b/c/"
	if size <= 0 || size > maxMessageSize {
		logging.LogAction("client", "message size is not valid (0 - 64K), defaulting to 64-byte size")
		size = 64
	}

	return mqtt.Publish{
		Header:  mqtt.Header{QOS: 0},
		Topic:   []byte(topic),
		Payload: make([]byte, size),
	}
}

// NewConn creates a new connection for the load test.
func newConn(hostAndPort, key, channel string) (*conn, error) {
	socket, err := dial("tcp", hostAndPort)
	if err != nil {
		return nil, err
	}

	cli := &conn{
		Conn:    socket,
		scratch: make([]byte, 1),
		topic:   fmt.Sprintf("%s/%s", key, channel),
	}
	cli.Skip(mqtt.TypeOfConnack)

	// Connect to the broker
	logging.LogTarget("client", "connecting to the broker", hostAndPort)
	connect := mqtt.Connect{ClientID: []byte("load-tester")}
	if _, err := connect.EncodeTo(cli); err != nil {
		return nil, err
	}

	// Subscribe to the topic
	sub := mqtt.Subscribe{
		Header: mqtt.Header{QOS: 0},
		Subscriptions: []mqtt.TopicQOSTuple{
			{Topic: []byte(cli.topic), Qos: 0},
		},
	}
	if _, err := sub.EncodeTo(cli); err != nil {
		return nil, err
	}

	logging.LogTarget("client", "subscribing to the channel", channel)
	return cli, nil
}

// Conn represents a connection to use for the load test.
type conn struct {
	net.Conn
	scratch []byte
	topic   string
}

// ReadByte reads a single byte.
func (c *conn) ReadByte() (byte, error) {
	if _, err := io.ReadFull(c.Conn, c.scratch); err != nil {
		return 0, err
	}
	return c.scratch[0], nil
}

// Drain continously drains the connection.
func (c *conn) Drain() {
	for {
		if _, err := io.Copy(ioutil.Discard, c.Conn); err != nil {
			return
		}
	}
}

// Skip skips a single message or returns an error if the message doesn't match.
func (c *conn) Skip(mqttType uint8) error {
	pkt, err := mqtt.DecodePacket(c, 65536)
	if err != nil {
		return err
	}

	if pkt.Type() != mqttType {
		return fmt.Errorf("mqtt type is %v instead of %v", pkt.Type(), mqttType)
	}
	return nil
}
