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
	"math/rand"
	"net"
	"testing"

	conf "github.com/emitter-io/config"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/provider/usage"
)

// 30000	     44945 ns/op	    1558 B/op	      21 allocs/op
func BenchmarkSerial(b *testing.B) {
	broker, cli := brokerAndClient(rand.Intn(10000) + 2000)
	defer broker.Close()
	defer cli.Close()

	// Connect to the broker
	connect := mqtt.Connect{ClientID: []byte("test")}
	check(connect.EncodeTo(cli))
	responseOf(mqtt.TypeOfConnack, cli)

	// Subscribe to a topic
	sub := mqtt.Subscribe{
		Header: &mqtt.StaticHeader{QOS: 0},
		Subscriptions: []mqtt.TopicQOSTuple{
			{Topic: []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"), Qos: 0},
		},
	}
	check(sub.EncodeTo(cli))
	responseOf(mqtt.TypeOfSuback, cli)

	// Prepare a message for the benchmark
	msg := mqtt.Publish{
		Header:  &mqtt.StaticHeader{QOS: 0},
		Topic:   []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"),
		Payload: []byte("hello world"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		check(msg.EncodeTo(cli))
		responseOf(mqtt.TypeOfPublish, cli)
	}
}

func responseOf(mqttType uint8, cli net.Conn) {
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

func brokerAndClient(port int) (*Service, net.Conn) {
	cfg := config.NewDefault().(*config.Config)
	cfg.License = testLicense
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

	// Create a client
	cli, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		broker.Close()
		panic(err)
	}
	return broker, cli
}
