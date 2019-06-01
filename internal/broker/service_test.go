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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	secmock "github.com/emitter-io/emitter/internal/provider/contract/mock"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security/license"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testLicense   = "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI"
	testLicenseV2 = "RfBEIIFz1nNLf12JYRpoEUqFPLb3na0X_xbP_h3PM_CqDUVBGJfEV3WalW2maauQd48o-TcTM_61BfEsELfk0qMDqrCTswkB:2"
)

func Test_onHTTPPresence(t *testing.T) {
	license, _ := license.Parse(testLicense)

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
	const port = 9996
	broker := newTestBroker(port, 1)
	defer broker.Close()

	cli := newTestClient(port)
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
		pkt, err := mqtt.DecodePacket(cli, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPingresp, pkt.Type())
	}

	{ // Publish a retained message
		msg := mqtt.Publish{
			Header:  mqtt.Header{QOS: 0, Retain: true},
			Topic:   []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"),
			Payload: []byte("retained message"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Subscribe to a topic
		sub := mqtt.Subscribe{
			Header: mqtt.Header{QOS: 0},
			Subscriptions: []mqtt.TopicQOSTuple{
				{Topic: []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"), Qos: 0},
			},
		}
		_, err := sub.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read the retained message
		pkt, err := mqtt.DecodePacket(cli, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
		assert.Equal(t, &mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte("a/b/c/"),
			Payload: []byte("retained message"),
		}, pkt)
	}

	{ // Read suback
		pkt, err := mqtt.DecodePacket(cli, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfSuback, pkt.Type())
	}

	{ // Publish a message
		msg := mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read the message back
		pkt, err := mqtt.DecodePacket(cli, 65536)
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
			Topic:   []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/?me=0"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Unsubscribe from the topic
		sub := mqtt.Unsubscribe{
			Header: mqtt.Header{QOS: 0},
			Topics: []mqtt.TopicQOSTuple{
				{Topic: []byte("EbUlduEbUssgWueAWjkEZwdYG5YC0dGh/a/b/c/"), Qos: 0},
			},
		}
		_, err := sub.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read unsuback
		pkt, err := mqtt.DecodePacket(cli, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfUnsuback, pkt.Type())
	}

	{ // Create a private link
		msg := mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
			Topic:   []byte("emitter/link/?req=1"),
			Payload: []byte(`{ "name": "hi", "key": "k44Ss59ZSxg6Zyz39kLwN-2t5AETnGpm", "channel": "a/b/c/", "private": true }`),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Read the link response
		pkt, err := mqtt.DecodePacket(cli, 65536)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.TypeOfPublish, pkt.Type())
	}

	{ // Publish a message to a link
		msg := mqtt.Publish{
			Header:  mqtt.Header{QOS: 0},
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
