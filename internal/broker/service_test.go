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
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/stretchr/testify/assert"
)

const (
	testLicense   = "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI"
	testLicenseV2 = "RfBEIIFz1nNLf12JYRpoEUqFPLb3na0X_xbP_h3PM_CqDUVBGJfEV3WalW2maauQd48o-TcTM_61BfEsELfk0qMDqrCTswkB:2"
)

func TestPubsub(t *testing.T) {
	const port = 9996
	broker := newTestBroker(port, 2)
	defer func() {
		time.Sleep(500 * time.Millisecond)
		broker.Close()
	}()

	c1, c2 := newTestClient(port), newTestClient(port)
	defer c1.Close()
	defer c2.Close()

	key1 := "w07Jv3TMhYTg6lLk6fQoVG2KCe7gjFPk" // on a/b/c/ with 'rwslp'

	{ // Connect to the broker (client1)
		connect := mqtt.Connect{
			ClientID:       []byte("test"),
			WillFlag:       true,
			WillRetainFlag: false,
			WillTopic:      []byte(key1 + "/a/b/c/"),
			WillMessage:    []byte("last will message"),
		}
		n, err := connect.EncodeTo(c1)
		assert.Equal(t, 74, n)
		assert.NoError(t, err)
	}

	{ // Read connack
		pkt, err := mqtt.DecodePacket(c1, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfConnack, pkt.Type())
	}

	{ // Connect to the broker (client2)
		connect := mqtt.Connect{
			ClientID: []byte("test2"),
		}
		n, err := connect.EncodeTo(c2)
		assert.Equal(t, 15, n)
		assert.NoError(t, err)
	}

	{ // Read connack
		pkt, err := mqtt.DecodePacket(c2, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfConnack, pkt.Type())
	}

	{ // Ping the broker
		ping := mqtt.Pingreq{}
		n, err := ping.EncodeTo(c1)
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
	}

	{ // Read pong
		pkt, err := mqtt.DecodePacket(c1, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPingresp, pkt.Type())
	}

	{ // Publish a retained message
		msg := mqtt.Publish{
			Header:  mqtt.Header{QOS: 0, Retain: true},
			Topic:   []byte(key1 + "/a/b/c/"),
			Payload: []byte("retained message"),
		}
		_, err := msg.EncodeTo(c1)
		assert.NoError(t, err)
	}

	{ // Subscribe to a topic
		sub := mqtt.Subscribe{
			Header: mqtt.Header{QOS: 0},
			Subscriptions: []mqtt.TopicQOSTuple{
				{Topic: []byte(key1 + "/a/b/c/"), Qos: 0},
			},
		}
		_, err := sub.EncodeTo(c1)
		assert.NoError(t, err)
	}

	{ // Read the retained message
		pkt, err := mqtt.DecodePacket(c1, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
		assert.Equal(t, &mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte("a/b/c/"),
			Payload: []byte("retained message"),
		}, pkt)
	}

	{ // Read suback
		pkt, err := mqtt.DecodePacket(c1, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfSuback, pkt.Type())
	}

	{ // Publish a message
		msg := mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte(key1 + "/a/b/c/"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(c1)
		assert.NoError(t, err)
	}

	{ // Read the message back
		pkt, err := mqtt.DecodePacket(c1, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
		assert.Equal(t, &mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte("a/b/c/"),
			Payload: []byte("hello world"),
		}, pkt)
	}

	{ // Publish a message but ignore ourselves
		msg := mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte(key1 + "/a/b/c/?me=0"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(c1)
		assert.NoError(t, err)
	}

	{ // Unsubscribe from the topic
		sub := mqtt.Unsubscribe{
			Header: mqtt.Header{QOS: 0},
			Topics: []mqtt.TopicQOSTuple{
				{Topic: []byte(key1 + "/a/b/c/"), Qos: 0},
			},
		}
		_, err := sub.EncodeTo(c1)
		assert.NoError(t, err)
	}

	{ // Read unsuback
		pkt, err := mqtt.DecodePacket(c1, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfUnsuback, pkt.Type())
	}

	{ // Create a private link
		msg := mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte("emitter/link/?req=1"),
			Payload: []byte(`{ "name": "hi", "key": "k44Ss59ZSxg6Zyz39kLwN-2t5AETnGpm", "channel": "a/b/c/", "private": true }`),
		}
		_, err := msg.EncodeTo(c1)
		assert.NoError(t, err)
	}

	{ // Read the link response
		pkt, err := mqtt.DecodePacket(c1, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
	}

	{ // Publish a message to a link
		msg := mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte("hi"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(c1)
		assert.NoError(t, err)
	}

	{ // Subscribe to a topic (client2), but do not read retained
		sub := mqtt.Subscribe{
			Header: mqtt.Header{QOS: 0},
			Subscriptions: []mqtt.TopicQOSTuple{
				{Topic: []byte(key1 + "/a/b/c/?last=0"), Qos: 0},
			},
		}
		_, err := sub.EncodeTo(c2)
		assert.NoError(t, err)
	}

	{ // Read suback
		pkt, err := mqtt.DecodePacket(c2, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfSuback, pkt.Type())
	}

	{ // Disconnect from the broker
		disconnect := mqtt.Disconnect{}
		n, err := disconnect.EncodeTo(c1)
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
	}

	{ // Wait to be closed
		_, err := mqtt.DecodePacket(c1, 65536)
		assert.NoError(t, err)
	}

	{ // Read last will
		pkt, err := mqtt.DecodePacket(c2, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
		assert.Equal(t, &mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte("a/b/c/"),
			Payload: []byte("last will message"),
		}, pkt)
	}

}
