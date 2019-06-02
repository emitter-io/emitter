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

package bench

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/emitter-io/emitter/internal/network/mqtt"
)

func main() {
	cli := newBenchClient(8080)
	println("draining...")
	go cli.Drain()
	msg := mqtt.Publish{
		Header:  &mqtt.StaticHeader{QOS: 0},
		Topic:   []byte("4kzJv3TMhYTg6lLk6fQoFG2KCe7gjFPk/a/b/c/"),
		Payload: []byte("hello world"),
	}

	for {
		for i := 0; i < 20; i++ {
			check(msg.EncodeTo(cli))
		}

		time.Sleep(1 * time.Millisecond)
	}
}

func responseOf(mqttType uint8, cli *testConn) {
	pkt, err := mqtt.DecodePacket(cli, 65536)
	if err != nil {
		panic(err)
	}
	if pkt.Type() != mqttType {
		panic(fmt.Errorf("mqtt type is %v instead of %v", pkt.Type(), mqttType))
	}
}

func check(_ int, err error) {
	if err != nil {
		panic(err)
	}
}

func newBenchClient(port int) *testConn {
	socket, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}
	cli := &testConn{
		Conn:    socket,
		buffer:  make([]byte, 8*1024),
		scratch: make([]byte, 1),
	}
	connect := mqtt.Connect{ClientID: []byte("test")}
	check(connect.EncodeTo(cli))

	// Subscribe to a topic
	sub := mqtt.Subscribe{
		Header: mqtt.Header{QOS: 0},
		Subscriptions: []mqtt.TopicQOSTuple{
			{Topic: []byte("4kzJv3TMhYTg6lLk6fQoFG2KCe7gjFPk/a/b/c/"), Qos: 0},
		},
	}
	check(sub.EncodeTo(cli))
	return cli
}

type testConn struct {
	net.Conn
	buffer  []byte
	scratch []byte
}

func (c *testConn) ReadByte() (byte, error) {
	if _, err := io.ReadFull(c.Conn, c.scratch); err != nil {
		return 0, err
	}
	return c.scratch[0], nil
}

func (c *testConn) Drain() {
	for {
		if _, err := io.Copy(ioutil.Discard, c.Conn); err != nil {
			return
		}
	}
}
