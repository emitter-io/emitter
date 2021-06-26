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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"testing"
	"time"

	conf "github.com/emitter-io/config"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/provider/usage"
)

var benchInit sync.Once

// BenchmarkSerial-8   	     100	  10007938 ns/op	    2172 B/op	      17 allocs/op
func BenchmarkSerial(b *testing.B) {
	const port = 9995
	benchInit.Do(func() {
		newTestBroker(port, 2)
	})

	// Prepare a message for the benchmark
	cli := newBenchClient(port)
	defer cli.Close()

	responseOf(mqtt.TypeOfConnack, cli)
	responseOf(mqtt.TypeOfSuback, cli)

	msg := mqtt.Publish{
		Header:  mqtt.Header{QOS: 0},
		Topic:   []byte("w07Jv3TMhYTg6lLk6fQoVG2KCe7gjFPk/a/b/c/"),
		Payload: []byte("hello world"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		// Since this is serial, we should get the same rate as the underlying
		// rate limiting on the socket.
		check(msg.EncodeTo(cli))
		responseOf(mqtt.TypeOfPublish, cli)
	}
}

// BenchmarkParallel-8   	  200000	      8068 ns/op	    1331 B/op	      13 allocs/op
func BenchmarkParallel(b *testing.B) {
	const port = 9995
	benchInit.Do(func() {
		newTestBroker(port, 2)
	})

	// Prepare a message for the benchmark
	cli := newBenchClient(port)
	defer cli.Close()
	msg := mqtt.Publish{
		Header:  mqtt.Header{QOS: 0},
		Topic:   []byte("w07Jv3TMhYTg6lLk6fQoVG2KCe7gjFPk/a/b/c/"),
		Payload: []byte("hello world"),
	}

	go cli.Drain()

	//defer profile.Start().Stop()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		check(msg.EncodeTo(cli))
	}
}

// BenchmarkFanOut/8-Clients-8         	  214063	      6407 ns/op	    1423 B/op	      19 allocs/op
// BenchmarkFanOut/16-Clients-8        	  299041	      4140 ns/op	     999 B/op	      13 allocs/op
// BenchmarkFanOut/32-Clients-8        	  280191	      5187 ns/op	    1425 B/op	       9 allocs/op
// BenchmarkFanOut/64-Clients-8        	  294547	     11597 ns/op	    3406 B/op	      12 allocs/op
// BenchmarkFanOut/128-Clients-8       	  231928	     19599 ns/op	    5788 B/op	      12 allocs/op
func BenchmarkFanOut(b *testing.B) {
	const port = 9995
	benchInit.Do(func() {
		newTestBroker(port, 2)
	})

	for n := 8; n <= 128; n *= 2 {
		b.Run(fmt.Sprintf("%d-Clients", n), func(b *testing.B) {
			var clients []*testConn
			for i := 0; i < n; i++ {
				cli := newBenchClient(port)
				clients = append(clients, cli)
			}

			// Prepare a message for the benchmark
			msg := mqtt.Publish{
				Header:  mqtt.Header{QOS: 0},
				Topic:   []byte("w07Jv3TMhYTg6lLk6fQoVG2KCe7gjFPk/a/b/c/"),
				Payload: []byte("hello world"),
			}

			for _, cli := range clients {
				cli := cli
				defer cli.Close()
				go cli.Drain()
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				check(msg.EncodeTo(clients[0]))
			}
		})
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
	cli := newTestClient(port)
	connect := mqtt.Connect{ClientID: []byte("test")}
	check(connect.EncodeTo(cli))

	// Subscribe to a topic
	sub := mqtt.Subscribe{
		Header: mqtt.Header{QOS: 0},
		Subscriptions: []mqtt.TopicQOSTuple{
			{Topic: []byte("w07Jv3TMhYTg6lLk6fQoVG2KCe7gjFPk/a/b/c/"), Qos: 0},
		},
	}
	check(sub.EncodeTo(cli))
	return cli
}

func newTestBroker(port int, licenseVersion int) *Service {
	cfg := config.NewDefault().(*config.Config)
	cfg.License = testLicense
	cfg.Debug = true
	if licenseVersion == 2 {
		cfg.License = testLicenseV2
	}

	cfg.TLS = &conf.TLSConfig{}
	cfg.ListenAddr = fmt.Sprintf("127.0.0.1:%d", port)
	cfg.Cluster = &config.ClusterConfig{
		NodeName:      "00:00:00:00:00:01",
		ListenAddr:    ":4000",
		AdvertiseAddr: ":4001",
	}

	// Start the broker asynchronously
	broker, err := NewService(context.Background(), cfg)
	if err != nil {
		broker.Close()
		panic(err)
	}

	broker.contracts = contract.NewSingleContractProvider(broker.License, usage.NewNoop())
	broker.storage = storage.NewInMemory(broker)
	broker.storage.Configure(nil)
	go broker.Listen()
	return broker
}

func newTestClient(port int) *testConn {
	var lastError error
	for i := 1; i <= 10; i++ {
		cli, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if lastError = err; lastError != nil {
			time.Sleep(time.Second)
			continue
		}

		return &testConn{
			Conn:    cli,
			buffer:  make([]byte, 8*1024),
			scratch: make([]byte, 1),
		}
	}

	panic(lastError)
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
