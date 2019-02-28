package broker

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	conf "github.com/emitter-io/config"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/contract"
	secmock "github.com/emitter-io/emitter/internal/provider/contract/mock"
	"github.com/emitter-io/emitter/internal/provider/storage"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testLicense = "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI"

func Test_onHTTPPresence(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)

	tests := []struct {
		payload       string
		contractValid bool
		contractFound bool
		status        int
		success       bool
		err           error
		resp          presenceResponse
		msg           string
	}{
		{
			payload:       `{"key":"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY","channel":"a","status":true}`,
			contractValid: true,
			contractFound: true,
			success:       true,
			status:        http.StatusOK,
			err:           nil,
			resp:          presenceResponse{Event: presenceStatusEvent, Channel: "a"},
			msg:           "Successful case",
		},
		{
			payload:       `{"key":"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY","channel":"a","status":true}`,
			contractValid: true,
			contractFound: true,
			success:       true,
			status:        http.StatusOK,
			err:           nil,
			resp:          presenceResponse{Event: presenceStatusEvent, Channel: "a"},
			msg:           "Successful case",
		},
		{
			payload:       "",
			err:           ErrBadRequest,
			success:       false,
			status:        http.StatusBadRequest,
			contractValid: true,
			contractFound: true,
			msg:           "Invalid payload case",
		},
		{
			payload:       `{"key":"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY","channel":"a+b","status":true}`,
			contractValid: true,
			contractFound: true,
			success:       false,
			status:        http.StatusBadRequest,
			err:           ErrBadRequest,
			msg:           "Invalid channel case",
		},
		{
			payload:       `{"key":"0Nq8SWbL8qoOKEDqh_ebBZRqJDby30m","channel":"a","status":true}`,
			contractValid: true,
			contractFound: true,
			success:       false,
			status:        http.StatusUnauthorized,
			err:           ErrUnauthorized,
			msg:           "Key for wrong channel case",
		},
		{
			payload:       `{"key":"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY","channel":"a+b","status":true}`,
			err:           ErrNotFound,
			status:        http.StatusNotFound,
			contractValid: true,
			contractFound: false,
			msg:           "Contract not found case",
		},
		{
			payload:       `{"key":"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY","channel":"a+b","status":true}`,
			err:           ErrUnauthorized,
			status:        http.StatusUnauthorized,
			contractValid: false,
			contractFound: true,
			msg:           "Contract is invalid case",
		},
	}

	for _, tc := range tests {

		contract := new(secmock.Contract)
		contract.On("Validate", mock.Anything).Return(tc.contractValid)
		contract.On("Stats").Return(usage.NewMeter(0))

		provider := secmock.NewContractProvider()
		provider.On("Get", mock.Anything).Return(contract, tc.contractFound)

		s := &Service{
			contracts:     provider,
			subscriptions: message.NewTrie(),
			License:       license,
		}
		s.Cipher, _ = s.License.Cipher()

		req, _ := http.NewRequest("POST", "/presence", strings.NewReader(tc.payload))

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.onHTTPPresence)
		handler.ServeHTTP(rr, req)

		var parsedResp presenceResponse
		json.Unmarshal(rr.Body.Bytes(), &parsedResp)

		// Check the response body is what we expect.
		assert.Equal(t, tc.status, rr.Code)
		assert.Equal(t, 0, len(parsedResp.Who))
	}
}

func TestPubsub(t *testing.T) {
	cfg := config.NewDefault().(*config.Config)
	cfg.License = testLicense
	cfg.ListenAddr = "127.0.0.1:9998"
	cfg.Cluster = nil
	cfg.TLS = &conf.TLSConfig{}

	// Start the broker asynchronously
	broker, svcErr := NewService(context.Background(), cfg)
	broker.contracts = contract.NewSingleContractProvider(broker.License, usage.NewNoop())
	broker.storage = storage.NewInMemory(broker)
	broker.storage.Configure(nil)
	assert.NoError(t, svcErr)
	defer broker.Close()
	go broker.Listen()

	// Create a client
	cli, dialErr := net.Dial("tcp", "127.0.0.1:9998")
	assert.NoError(t, dialErr)
	defer cli.Close()

	{ // Connect to the broker
		connect := mqtt.Connect{ClientID: []byte("test")}
		n, err := connect.EncodeTo(cli)
		assert.Equal(t, 14, n)
		assert.NoError(t, err)
	}

	{ // Read connack
		pkt, err := mqtt.DecodePacket(cli, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfConnack, pkt.Type())
	}

	{ // Ping the broker
		ping := mqtt.Pingreq{}
		n, err := ping.EncodeTo(cli)
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
	}

	{ // Read pong
		pkt, err := mqtt.DecodePacket(cli,65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPingresp, pkt.Type())
	}

	{ // Publish a retained message
		msg := mqtt.Publish{
			Header:  &mqtt.StaticHeader{QOS: 0, Retain: true},
			Topic:   []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"),
			Payload: []byte("retained message"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Subscribe to a topic
		sub := mqtt.Subscribe{
			Header: &mqtt.StaticHeader{QOS: 0},
			Subscriptions: []mqtt.TopicQOSTuple{
				{Topic: []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"), Qos: 0},
			},
		}
		_, err := sub.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read the retained message
		pkt, err := mqtt.DecodePacket(cli,65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
		assert.Equal(t, &mqtt.Publish{
			Header:  &mqtt.StaticHeader{QOS: 0},
			Topic:   []byte("a/b/c/"),
			Payload: []byte("retained message"),
		}, pkt)
	}

	{ // Read suback
		pkt, err := mqtt.DecodePacket(cli,65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfSuback, pkt.Type())
	}

	{ // Publish a message
		msg := mqtt.Publish{
			Header:  &mqtt.StaticHeader{QOS: 0},
			Topic:   []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read the message back
		pkt, err := mqtt.DecodePacket(cli,65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
		assert.Equal(t, &mqtt.Publish{
			Header:  &mqtt.StaticHeader{QOS: 0},
			Topic:   []byte("a/b/c/"),
			Payload: []byte("hello world"),
		}, pkt)
	}

	{ // Publish a message but ignore ourselves
		msg := mqtt.Publish{
			Header:  &mqtt.StaticHeader{QOS: 0},
			Topic:   []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/?me=0"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Unsubscribe from the topic
		sub := mqtt.Unsubscribe{
			Header: &mqtt.StaticHeader{QOS: 0},
			Topics: []mqtt.TopicQOSTuple{
				{Topic: []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"), Qos: 0},
			},
		}
		_, err := sub.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read unsuback
		pkt, err := mqtt.DecodePacket(cli,65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfUnsuback, pkt.Type())
	}

	{ // Create a private link
		msg := mqtt.Publish{
			Header:  &mqtt.StaticHeader{QOS: 0},
			Topic:   []byte("emitter/link/?req=1"),
			Payload: []byte(`{ "name": "hi", "key": "k44Ss59ZSxg6Zyz39kLwN-2t5AETnGpm", "channel": "a/b/c/", "private": true }`),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read the link response
		pkt, err := mqtt.DecodePacket(cli,65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
	}

	{ // Publish a message to a link
		msg := mqtt.Publish{
			Header:  &mqtt.StaticHeader{QOS: 0},
			Topic:   []byte("hi"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Disconnect from the broker
		disconnect := mqtt.Disconnect{}
		n, err := disconnect.EncodeTo(cli)
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
	}

}
