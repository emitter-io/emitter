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
	"testing"

	"github.com/emitter-io/emitter/internal/service/cluster"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	netmock "github.com/emitter-io/emitter/internal/network/mock"
	"github.com/emitter-io/emitter/internal/network/mqtt"
	"github.com/emitter-io/emitter/internal/provider/contract"
	secmock "github.com/emitter-io/emitter/internal/provider/contract/mock"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/security/license"
	"github.com/emitter-io/emitter/internal/service/keygen"
	"github.com/emitter-io/stats"
	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_onLink(t *testing.T) {
	tests := []struct {
		packet  string
		channel string
		success bool
	}{
		{
			packet:  `{ "name": "AB", "key": "k44Ss59ZSxg6Zyz39kLwN-2t5AETnGpm", "channel": "a/b/c/", "subscribe": true }`,
			channel: "a/b/c/",
			success: true,
		},
		{
			packet:  `{ "name": "AB", "key": "k44Ss59ZSxg6Zyz39kLwN-2t5AETnGpm", "channel": "a/b/c/"}`,
			channel: "a/b/c/",
			success: true,
		},
		{
			packet:  `{ "name": "AB", "key": "xxx", "channel": "a/b/c/", "subscribe": true }`,
			channel: "a/b/c/",
			success: true,
		},
		{packet: `{ "name": "ABC", "key": "k44Ss59ZSxg6Zyz39kLwN-2t5AETnGpm", "channel": "a/b/c/", "subscribe": true }`},
		{packet: `{ "name": "", "key": "k44Ss59ZSxg6Zyz39kLwN-2t5AETnGpm", "channel": "a/b/c/", "subscribe": true }`},
		{packet: `{"key": "k44Ss59ZSxg6Zyz39kLwN-2t5AETnGpm", "channel": "a/b/c/",  "subscribe": true }`},
		{packet: `{ "name": "AB", "key": "k44Ss59ZSxg6Zyz39kLwN-2t5AETnGpm", "channel": "---", "subscribe": true }`},
	}

	for _, tc := range tests {
		t.Run(tc.packet, func(*testing.T) {
			provider := secmock.NewContractProvider()
			contract := new(secmock.Contract)
			contract.On("Validate", mock.Anything).Return(true)
			provider.On("Get", mock.Anything).Return(contract, true)
			license, _ := license.Parse(testLicense)
			cipher, _ := license.Cipher()
			s := &Service{
				contracts:     provider,
				subscriptions: message.NewTrie(),
				License:       license,
				presence:      make(chan *presenceNotify, 100),
			}

			s.Keygen = keygen.NewProvider(cipher, provider, s)
			conn := netmock.NewConn()
			nc := s.newConn(conn.Client, 0)

			resp, ok := nc.onLink([]byte(tc.packet))
			assert.Equal(t, tc.success, ok)
			if tc.success {
				assert.Contains(t, resp.(*linkResponse).Channel, tc.channel)
			}
		})
	}
}

func TestHandlers_onMe(t *testing.T) {
	license, _ := license.Parse(testLicense)
	s := &Service{
		subscriptions: message.NewTrie(),
		License:       license,
	}

	conn := netmock.NewConn()
	nc := s.newConn(conn.Client, 0)
	nc.links["0"] = "key/a/b/c/"
	resp, success := nc.onMe()
	meResp := resp.(*meResponse)

	assert.True(t, success)
	assert.Equal(t, "a/b/c/", meResp.Links["0"])
	assert.NotNil(t, resp)
	assert.NotZero(t, len(meResp.ID))
}

func TestHandlers_onSubscribeUnsubscribe(t *testing.T) {
	license, _ := license.Parse(testLicense)
	tests := []struct {
		channel       string
		subCount      int
		subErr        error
		unsubCount    int
		unsubErr      error
		contractValid bool
		contractFound bool
		msg           string
	}{
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			subCount:      1,
			subErr:        (*errors.Error)(nil),
			unsubCount:    0,
			unsubErr:      (*errors.Error)(nil),
			contractValid: true,
			contractFound: true,
			msg:           "Successful case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a+q/b/c/",
			subCount:      0,
			subErr:        errors.ErrBadRequest,
			unsubCount:    0,
			unsubErr:      errors.ErrBadRequest,
			contractValid: true,
			contractFound: true,
			msg:           "Invalid channel case",
		}, /*
			{
				channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/+/b/c/",
				count:         0,
				err:           ErrForbidden,
				contractValid: true,
				contractFound: true,
				msg:           "Channel is not static case",
			},*/

		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBZRqJDby30mT/a/b/c/",
			subCount:      0,
			subErr:        errors.ErrUnauthorized,
			unsubCount:    0,
			unsubErr:      errors.ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Expired key case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			subCount:      0,
			subErr:        errors.ErrUnauthorized,
			unsubCount:    0,
			unsubErr:      errors.ErrUnauthorized,
			contractValid: true,
			contractFound: false,
			msg:           "Contract not found case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			subCount:      0,
			subErr:        errors.ErrUnauthorized,
			unsubCount:    0,
			unsubErr:      errors.ErrUnauthorized,
			contractValid: false,
			contractFound: true,
			msg:           "Contract is invalid case",
		}, /*
			{
				channel:       "0Nq8SWbL8qoJzie4_C4yvupug6cLLlWO/a/b/c/",
				count:         0,
				err:           ErrUnauthorized,
				contractValid: true,
				contractFound: true,
				msg:           "No write permission case",
			},*/
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBZHmCtcvoHGQ/a/b/c/",
			subCount:      0,
			subErr:        errors.ErrUnauthorized,
			unsubCount:    0,
			unsubErr:      errors.ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Wrong target case",
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(*testing.T) {
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
				presence:      make(chan *presenceNotify, 100),
			}

			s.Keygen = keygen.NewProvider(cipher, provider, s)
			conn := netmock.NewConn()
			nc := s.newConn(conn.Client, 0)

			// Subscribe and check for error.
			subErr := nc.onSubscribe([]byte(tc.channel))
			assert.Equal(t, tc.subErr, subErr, tc.msg)

			// Search for the ssid.
			channel := security.ParseChannel([]byte(tc.channel))
			key, _ := cipher.DecryptKey(channel.Key)
			ssid := message.NewSsid(key.Contract(), channel.Query)
			subscribers := s.subscriptions.Lookup(ssid, nil)
			assert.Equal(t, tc.subCount, len(subscribers))

			// Unsubscribe and check for error.
			unsubErr := nc.onUnsubscribe([]byte(tc.channel))
			assert.Equal(t, tc.unsubErr, unsubErr, tc.msg)

			// Search for the ssid.
			subscribers = s.subscriptions.Lookup(ssid, nil)
			assert.Equal(t, tc.unsubCount, len(subscribers))
		})
	}
}

func TestHandlers_onPublish(t *testing.T) {
	license, _ := license.Parse(testLicense)
	tests := []struct {
		channel       string
		payload       string
		err           error
		contractValid bool
		contractFound bool
		msg           string
	}{
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			payload:       "test",
			err:           (*errors.Error)(nil),
			contractValid: true,
			contractFound: true,
			msg:           "Successful case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a+q/b/c/",
			payload:       "test",
			err:           errors.ErrBadRequest,
			contractValid: true,
			contractFound: true,
			msg:           "Invalid channel case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/+/b/c/",
			payload:       "test",
			err:           errors.ErrForbidden,
			contractValid: true,
			contractFound: true,
			msg:           "Channel is not static case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBZRqJDby30mT/a/b/c/",
			payload:       "test",
			err:           errors.ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Expired key case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			payload:       "test",
			err:           errors.ErrUnauthorized,
			contractValid: true,
			contractFound: false,
			msg:           "Contract not found case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			payload:       "test",
			err:           errors.ErrUnauthorized,
			contractValid: false,
			contractFound: true,
			msg:           "Contract is invalid case",
		},
		{
			channel:       "0Nq8SWbL8qoJzie4_C4yvupug6cLLlWO/a/b/c/",
			payload:       "test",
			err:           errors.ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "No write permission case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBZHmCtcvoHGQ/a/b/c/",
			payload:       "test",
			err:           errors.ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Wrong target case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			payload:       "test",
			err:           (*errors.Error)(nil),
			contractValid: true,
			contractFound: true,
			msg:           "No store permission case",
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
		}

		s.Keygen = keygen.NewProvider(cipher, provider, s)
		conn := netmock.NewConn()
		nc := s.newConn(conn.Client, 0)

		err := nc.onPublish(&mqtt.Publish{
			Topic:   []byte(tc.channel),
			Payload: []byte(tc.payload),
		})

		assert.Equal(t, tc.err, err, tc.msg)
	}
}

func TestHandlers_onPresence(t *testing.T) {
	// TODO :
	// - valid key for the right channel, but no presence right.
	// - test Who
	license, _ := license.Parse(testLicenseV2)
	tests := []struct {
		channel       string
		payload       string
		contractValid bool
		contractFound bool
		success       bool
		err           error
		resp          presenceResponse
		msg           string
	}{
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"hw7Jv3TMhYTg6lLk2fQoSvs2EP3gjFPk\",\"channel\":\"a\",\"status\":true}",
			contractValid: true,
			contractFound: true,
			success:       true,
			err:           nil,
			resp:          presenceResponse{Event: presenceStatusEvent, Channel: "a"},
			msg:           "Successful case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "",
			err:           errors.ErrBadRequest,
			success:       false,
			contractValid: true,
			contractFound: true,
			msg:           "Invalid payload case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"hw7Jv3TMhYTg6lLk2fQoSvs2EP3gjFPk\",\"channel\":\"a+b\",\"status\":true}",
			contractValid: true,
			contractFound: true,
			success:       false,
			err:           errors.ErrBadRequest,
			msg:           "Invalid channel case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"07XJv3TMhYTg6lLk2fQoSift1AbgjFPk\",\"channel\":\"a\",\"status\":true}",
			contractValid: true,
			contractFound: true,
			success:       false,
			err:           errors.ErrUnauthorized,
			msg:           "Key for wrong channel case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"hw7Jv3TMhYTg6lLk2fQoSvs2EP3gjFPk\",\"channel\":\"a\",\"status\":true}",
			err:           errors.ErrUnauthorized,
			contractValid: true,
			contractFound: false,
			msg:           "Contract not found case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"hw7Jv3TMhYTg6lLk2fQoSvs2EP3gjFPk\",\"channel\":\"a\",\"status\":true}",
			err:           errors.ErrUnauthorized,
			contractValid: false,
			contractFound: true,
			msg:           "Contract is invalid case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"sVTJv3TMhYTg6lLk2fQoCvs2EP3gjFPk\",\"channel\":\"a\",\"status\":true}",
			err:           errors.ErrUnauthorizedExt,
			contractValid: true,
			contractFound: true,
			msg:           "Extended key is unauthorized case",
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
		}

		s.Keygen = keygen.NewProvider(cipher, provider, s)
		conn := netmock.NewConn()
		nc := s.newConn(conn.Client, 0)

		resp, success := nc.onPresence([]byte(tc.payload))

		assert.Equal(t, tc.success, success, tc.msg)
		if !success {
			assert.Equal(t, tc.err, resp, tc.msg)
		}
	}
}

func TestHandlers_onKeygen(t *testing.T) {
	license, _ := license.Parse("N7XxQbUEPxJ_RIj4muLUdLGYtR1kdKe2AAAAAAAAAAI")
	tests := []struct {
		payload       string
		contractValid bool
		contractFound bool
		generated     bool
		resp          interface{}
		msg           string
	}{
		{
			payload:       "+{\"key\":\"xEbaDPaICEwVhgdnl2rg_1DWi_MAg_3B\",\"channel\":\"article1\"}",
			contractValid: true,
			contractFound: true,
			generated:     false,
			resp:          errors.ErrBadRequest,
			msg:           "Invalid request case",
		},
		{
			payload:       "{\"key\":\"xEbaDPaICEwVhgdnl2rg_1DWi_MAg_3B\",\"channel\":\"article1\"}",
			contractValid: true,
			contractFound: true,
			generated:     false,
			resp:          errors.ErrUnauthorized,
			msg:           "No keygen permission case (not a master key)",
		},
		{
			payload:       "{\"key\":\"8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR\",\"channel\":\"article1\"}",
			contractValid: true,
			contractFound: false,
			generated:     false,
			resp:          errors.ErrNotFound,
			msg:           "Contract not found case",
		},
		{
			payload:       "{\"key\":\"8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR\",\"channel\":\"article1\"}",
			contractValid: false,
			contractFound: true,
			generated:     false,
			resp:          errors.ErrUnauthorized,
			msg:           "Contract not valid case",
		},
		{
			payload:       "{\"key\":\"8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR\",\"channel\":\"article1\"}",
			contractValid: true,
			contractFound: true,
			generated:     false,
			resp:          errors.ErrTargetInvalid,
			msg:           "Target invalid case",
		},
		{
			payload:       "{\"key\":\"8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR\",\"channel\":\"article1/\"}",
			contractValid: true,
			contractFound: true,
			generated:     true,
			resp:          keyGenResponse{Status: 200, Key: "76w5HdpyIOQh70HnB4d33gbqD5fFztGY", Channel: "article1/"},
			msg:           "Successful case",
		},
	}

	//keyGenResponse{Status: 200, Key: "76w5HdpyIOQh70HnB4d33gbqD5fFztGY", Channel: "article1"},
	for _, tc := range tests {
		t.Run(tc.msg, func(*testing.T) {
			provider := secmock.NewContractProvider()
			contract := new(secmock.Contract)
			contract.On("Validate", mock.Anything).Return(tc.contractValid)
			contract.On("Stats").Return(usage.NewMeter(0))
			provider.On("Get", mock.Anything).Return(contract, tc.contractFound)
			cipher, _ := license.Cipher()
			s := &Service{
				contracts:     provider,
				subscriptions: message.NewTrie(),
				License:       license,
			}

			s.Keygen = keygen.NewProvider(cipher, provider, s)
			conn := netmock.NewConn()
			nc := s.newConn(conn.Client, 0)

			//resp
			resp, generated := nc.onKeyGen([]byte(tc.payload))
			assert.Equal(t, tc.generated, generated, tc.msg)

			if !generated {
				keyGenResp := resp.(*errors.Error)
				assert.Equal(t, tc.resp, keyGenResp)
			} else {
				keyGenResp := resp.(*keyGenResponse)
				expected := tc.resp.(keyGenResponse)
				//assert.Equal(t, tc.resp, keyGenResp)
				assert.Equal(t, expected.Status, keyGenResp.Status)
			}
		})
	}
}

func TestHandlers_onEmitterRequest(t *testing.T) {
	tests := []struct {
		channel string
		payload string
		query   []uint32
		success bool
	}{
		{
			channel: "wrong",
			success: false,
		},
		{
			channel: "wrong",
			query:   []uint32{1, 2, 3},
			success: false,
		},
		{
			channel: "keygen",
			query:   []uint32{requestKeygen},
			success: false,
		},
		{
			channel: "presence",
			query:   []uint32{requestPresence},
			success: false,
		},
		{
			channel: "me",
			query:   []uint32{requestMe},
			success: true,
		},
		{
			channel: "link",
			query:   []uint32{requestLink},
			success: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.channel, func(*testing.T) {
			channel := &security.Channel{
				Key:     []byte("emitter"),
				Channel: []byte(tc.channel),
				Query:   tc.query,
			}

			s := &Service{
				contracts:     contract.NewNoopContractProvider(),
				subscriptions: message.NewTrie(),
				measurer:      stats.NewNoop(),
			}

			nc := s.newConn(netmock.NewNoop(), 0)
			ok := nc.onEmitterRequest(channel, []byte(tc.payload), 0)
			assert.Equal(t, tc.success, ok, tc.channel)
		})
	}
}

func TestHandlers_OnSurvey(t *testing.T) {
	encode := func(ssid ...uint32) []byte { b, _ := binary.Marshal(ssid); return b }
	tests := []struct {
		queryType string
		payload   []byte
		success   bool
	}{
		{
			queryType: "wrong",
			success:   false,
		},
		{
			queryType: "presence",
			payload:   []byte("hi"),
			success:   false,
		},
		{
			queryType: "presence",
			payload:   encode(1, 2, 3),
			success:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.queryType, func(*testing.T) {
			s := &Service{
				contracts:     contract.NewNoopContractProvider(),
				subscriptions: message.NewTrie(),
			}

			_, ok := s.OnSurvey(tc.queryType, tc.payload)
			assert.Equal(t, tc.success, ok, tc.queryType)
		})
	}
}

func TestHandlers_lookupPresence(t *testing.T) {
	s := &Service{
		contracts:     contract.NewNoopContractProvider(),
		subscriptions: message.NewTrie(),
		measurer:      stats.NewNoop(),
	}

	s.subscriptions.Subscribe(message.Ssid{1, 2, 3}, s.newConn(netmock.NewNoop(), 0))
	presence := s.lookupPresence(message.Ssid{1, 2, 3})
	assert.NotEmpty(t, presence)
}

func TestHandlers_onKeyBan(t *testing.T) {
	license, _ := license.Parse(testLicenseV2)
	cipher, _ := license.Cipher()
	contract := new(secmock.Contract)
	contract.On("Validate", mock.Anything).Return(true)
	provider := secmock.NewContractProvider()
	provider.On("Get", mock.Anything).Return(contract, true)

	s := &Service{
		License: license,
		cluster: cluster.NewSwarm(&config.ClusterConfig{
			NodeName:      "00:00:00:00:00:01",
			ListenAddr:    ":4000",
			AdvertiseAddr: ":4001",
		}),
	}
	s.Keygen = keygen.NewProvider(cipher, provider, s)

	// Key should be allowed
	ev := event.Ban("6ijJv3TMhYTg6lLk2fQoVNbGrujgjFPk")

	// Issue a request to ban the key
	req, _ := json.Marshal(&keyBanRequest{
		Secret: "wnLJv3TMhYTg6lLkGfQoazo1-k7gjFPk",
		Target: string(ev),
		Banned: true,
	})
	nc := s.newConn(netmock.NewConn().Client, 0)
	r, ok := nc.onKeyBan(req)

	// Key should be banned now
	assert.True(t, ok)
	assert.Equal(t, &keyBanResponse{Status: 200, Banned: true}, r.(*keyBanResponse))
	assert.True(t, s.cluster.Contains(&ev))
}
