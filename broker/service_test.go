package broker

/*
import (
	"net"
	"testing"

	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/network/mqtt"
	"github.com/emitter-io/emitter/security"
	"github.com/stretchr/testify/assert"
)
*/
const testLicense = "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI"

/*
func TestPubsub(t *testing.T) {
	cfg := config.NewDefault()
	cfg.License = testLicense
	cfg.TCPPort = ":9998"
	cfg.TLSPort = ":9999"
	cfg.Cluster = nil

	// Start the broker asynchronously
	broker, svcErr := NewService(cfg)
	broker.ContractProvider = security.NewSingleContractProvider(broker.License)
	assert.NoError(t, svcErr)
	defer close(broker.Closing)
	go broker.Listen()

	// Create a client
	cli, dialErr := net.Dial("tcp", ":9998")
	assert.NoError(t, dialErr)
	defer cli.Close()

	{ // Connect to the broker
		connect := mqtt.Connect{ClientID: []byte("test")}
		n, err := connect.EncodeTo(cli)
		assert.Equal(t, 14, n)
		assert.NoError(t, err)
	}

	{ // Read connack
		pkt, err := mqtt.DecodePacket(cli)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfConnack, pkt.Type())
	}

	{ // Subscribe to a topic
		sub := mqtt.Subscribe{
			Header: &mqtt.StaticHeader{QOS: 0},
			Subscriptions: []mqtt.TopicQOSTuple{
				{Topic: []byte("0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/"), Qos: 0},
			},
		}
		_, err := sub.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read suback
		pkt, err := mqtt.DecodePacket(cli)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfSuback, pkt.Type())
	}

	{ // Publish a message
		msg := mqtt.Publish{
			Header:  &mqtt.StaticHeader{QOS: 0},
			Topic:   []byte("0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read the message back
		pkt, err := mqtt.DecodePacket(cli)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
		assert.Equal(t, &mqtt.Publish{
			Header:  &mqtt.StaticHeader{QOS: 0},
			Topic:   []byte("a/b/c/"),
			Payload: []byte("hello world"),
		}, pkt)
	}
}
*/
