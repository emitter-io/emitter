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
	"net"
	"sync"
	"testing"

	conf "github.com/emitter-io/config"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/provider/usage"
)

var benchInit sync.Once

// BenchmarkSerial-8          30000             44349 ns/op            1574 B/op         19 allocs/op
func BenchmarkSerial(b *testing.B) {
	const port = 9995
	benchInit.Do(func() {
		newTestBroker(port, 2)
	})

	// Prepare a message for the benchmark
	cli := newBenchClient(port)
	defer cli.Close()

	msg := mqtt.Publish{
		Header:  mqtt.Header{QOS: 0},
		Topic:   []byte("4kzJv3TMhYTg6lLk6fQoFG2KCe7gjFPk/a/b/c/"),
		Payload: []byte("hello world"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		check(msg.EncodeTo(cli))
		responseOf(mqtt.TypeOfPublish, cli)
	}
}

// BenchmarkParallel-8   	  300000	     12459 ns/op	     919 B/op	      10 allocs/op
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
		Topic:   []byte("4kzJv3TMhYTg6lLk6fQoFG2KCe7gjFPk/a/b/c/"),
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

// BenchmarkFanOut/8-Clients-8                10000            160866 ns/op            2505 B/op         48 allocs/op
// BenchmarkFanOut/16-Clients-8                5000            278455 ns/op            3488 B/op         80 allocs/op
// BenchmarkFanOut/32-Clients-8                2000            552022 ns/op            6906 B/op        146 allocs/op
// BenchmarkFanOut/64-Clients-8                1000           1029245 ns/op           13961 B/op        278 allocs/op
// BenchmarkFanOut/128-Clients-8               1000           1947789 ns/op           28034 B/op        538 allocs/op
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
				Topic:   []byte("4kzJv3TMhYTg6lLk6fQoFG2KCe7gjFPk/a/b/c/"),
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

				wg := new(sync.WaitGroup)
				wg.Add(len(clients))
				for _, cli := range clients {
					cli := cli
					go func() {
						responseOf(mqtt.TypeOfPublish, cli)
						wg.Done()
					}()
				}
				wg.Wait()
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
	responseOf(mqtt.TypeOfConnack, cli)

	// Subscribe to a topic
	sub := mqtt.Subscribe{
		Header: mqtt.Header{QOS: 0},
		Subscriptions: []mqtt.TopicQOSTuple{
			{Topic: []byte("4kzJv3TMhYTg6lLk6fQoFG2KCe7gjFPk/a/b/c/"), Qos: 0},
		},
	}
	check(sub.EncodeTo(cli))
	responseOf(mqtt.TypeOfSuback, cli)
	return cli
}

func newTestBroker(port int, licenseVersion int) *Service {
	cfg := config.NewDefault().(*config.Config)
	cfg.License = testLicense
	if licenseVersion == 2 {
		cfg.License = testLicenseV2
	}

	cfg.ListenAddr = fmt.Sprintf("127.0.0.1:%d", port)
	cfg.Cluster = nil
	cfg.TLS = &conf.TLSConfig{}

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
	cli, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}
	return &testConn{
		Conn:    cli,
		buffer:  make([]byte, 8*1024),
		scratch: make([]byte, 1),
	}
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
		if _, err := io.ReadFull(c.Conn, c.buffer); err != nil {
			return
		}
	}
}
