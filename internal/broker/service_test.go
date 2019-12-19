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
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/emitter-io/emitter/internal/broker/keygen"
	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/contract"
	secmock "github.com/emitter-io/emitter/internal/provider/contract/mock"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security/license"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testLicense       = "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI"
	testLicenseV2     = "RfBEIIFz1nNLf12JYRpoEUqFPLb3na0X_xbP_h3PM_CqDUVBGJfEV3WalW2maauQd48o-TcTM_61BfEsELfk0qMDqrCTswkB:2"
	keygenTestLicense = "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI:1"
	keygenTestSecret  = "kBCZch5re3Ue-kpG1Aa8Vo7BYvXZ3UwR"
)

func newKeygenProvider(t *testing.T) *keygen.Provider {
	l, err := license.Parse(keygenTestLicense)
	assert.NoError(t, err)

	cipher, err := l.Cipher()
	assert.NoError(t, err)

	return keygen.NewProvider(cipher, contract.NewSingleContractProvider(l, usage.NewNoop()))
}

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
			err:           errors.ErrBadRequest,
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
			err:           errors.ErrBadRequest,
			msg:           "Invalid channel case",
		},
		{
			payload:       `{"key":"0Nq8SWbL8qoOKEDqh_ebBZRqJDby30m","channel":"a","status":true}`,
			contractValid: true,
			contractFound: true,
			success:       false,
			status:        http.StatusUnauthorized,
			err:           errors.ErrUnauthorized,
			msg:           "Key for wrong channel case",
		},
		{
			payload:       `{"key":"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY","channel":"a+b","status":true}`,
			err:           errors.ErrNotFound,
			status:        http.StatusNotFound,
			contractValid: true,
			contractFound: false,
			msg:           "Contract not found case",
		},
		{
			payload:       `{"key":"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY","channel":"a+b","status":true}`,
			err:           errors.ErrUnauthorized,
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

		cipher, _ := license.Cipher()
		s := &Service{
			contracts:     provider,
			subscriptions: message.NewTrie(),
			License:       license,
			Keygen:        keygen.NewProvider(cipher, provider),
		}

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
	broker := newTestBroker(port, 2)
	defer broker.Close()

	cli := newTestClient(port)
	defer cli.Close()

	key1 := "w07Jv3TMhYTg6lLk6fQoVG2KCe7gjFPk" // on a/b/c/ with 'rwslp'

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
			Topic:   []byte(key1 + "/a/b/c/"),
			Payload: []byte("retained message"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Subscribe to a topic
		sub := mqtt.Subscribe{
			Header: mqtt.Header{QOS: 0},
			Subscriptions: []mqtt.TopicQOSTuple{
				{Topic: []byte(key1 + "/a/b/c/"), Qos: 0},
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
			Topic:   []byte(key1 + "/a/b/c/"),
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
			Topic:   []byte(key1 + "/a/b/c/?me=0"),
			Payload: []byte("hello world"),
		}
		_, err := msg.EncodeTo(cli)
		assert.NoError(t, err)
	}

	{ // Unsubscribe from the topic
		sub := mqtt.Unsubscribe{
			Header: mqtt.Header{QOS: 0},
			Topics: []mqtt.TopicQOSTuple{
				{Topic: []byte(key1 + "/a/b/c/"), Qos: 0},
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

func getChannelKey(t *testing.T, channel string) string {
	p := newKeygenProvider(t)
	keyGenHandler := p.HTTPJson()

	data := url.Values{}
	data.Set("key", keygenTestSecret)
	data.Set("channel", channel)
	data.Set("ttl", "300")
	data.Set("sub", "on")
	data.Set("pub", "on")
	data.Set("store", "on")
	data.Set("load", "on")
	data.Set("presence", "on")
	data.Set("extend", "off")
	req, _ := http.NewRequest("POST", "https://emitter.io/keygen_json", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	w := httptest.NewRecorder()

	// act
	keyGenHandler(w, req)
	content, err := ioutil.ReadAll(w.Body)
	if err != nil {
		log.Fatal(err)
		return ""
	}

	jsonRet := httpJsonResponse{Code: HttpJsonRetFail, Message: ""}
	err = json.Unmarshal(content, &jsonRet)

	if err != nil {
		log.Fatal(err)
		return ""
	}

	if jsonRet.Code != HttpJsonRetSuccess {
		return ""
	}

	return jsonRet.Message
}

func Test_onHTTPPublishJson(t *testing.T) {
	channelName := "bar/"
	channelKey := getChannelKey(t, channelName)

	if channelKey == "" {
		log.Fatal("generate channel key fail")
		return
	}

	type testCase struct {
		Scenario                 string
		Key                      string
		Channel                  string
		TTL                      string
		Message                  string
		ExpectedResponseContains string
		Code                     int
	}
	testCases := []testCase{
		{
			Scenario:                 "Request with valid arguments",
			Key:                      channelKey,
			Channel:                  channelName,
			TTL:                      "0",
			Message:                  "hello world",
			ExpectedResponseContains: "\"code\":0",
			Code:                     HttpJsonRetSuccess,
		},
		{
			Scenario:                 "Request with invalid arguments",
			Key:                      channelKey + strconv.Itoa(rand.Intn(100000000000)),
			Channel:                  channelName,
			TTL:                      "0",
			Message:                  "hello world",
			ExpectedResponseContains: "\"code\":0",
			Code:                     HttpJsonRetFail,
		},
		{
			Scenario:                 "Request with empty key",
			Key:                      "",
			Channel:                  channelName,
			TTL:                      "0",
			Message:                  "hello world",
			ExpectedResponseContains: "\"code\":0",
			Code:                     HttpJsonRetFail,
		},
		{
			Scenario:                 "Request with empty channel",
			Key:                      channelKey,
			Channel:                  "",
			TTL:                      "0",
			Message:                  "hello world",
			ExpectedResponseContains: "\"code\":0",
			Code:                     HttpJsonRetFail,
		},
		{
			Scenario:                 "Request with empty channel",
			Key:                      channelKey,
			Channel:                  channelName,
			TTL:                      "0",
			Message:                  "",
			ExpectedResponseContains: "\"code\":0",
			Code:                     HttpJsonRetSuccess,
		},
	}

	license, _ := license.Parse(testLicense)

	for _, c := range testCases {

		contract := new(secmock.Contract)
		contract.On("Validate", mock.Anything).Return(true)
		contract.On("Stats").Return(usage.NewMeter(0))

		provider := secmock.NewContractProvider()
		provider.On("Get", mock.Anything).Return(contract, true)

		cipher, _ := license.Cipher()
		s := &Service{
			contracts:     provider,
			subscriptions: message.NewTrie(),
			License:       license,
			Keygen:        keygen.NewProvider(cipher, provider),
		}

		messageMap := map[string]interface{}{"channel": c.Channel, "key": c.Key, "ttl": c.TTL, "message": c.Message}
		msgBytes, _ := json.Marshal(messageMap)

		req, _ := http.NewRequest("POST", "/publish_json", strings.NewReader(string(msgBytes)))

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.onHTTPPublishJson)
		handler.ServeHTTP(rr, req)

		respJson := httpJsonResponse{Code: HttpJsonRetFail, Message: ""}
		json.Unmarshal(rr.Body.Bytes(), &respJson)

		// Check the response body is what we expect.
		assert.Equal(t, c.Code, respJson.Code, c.Scenario)
	}
}
