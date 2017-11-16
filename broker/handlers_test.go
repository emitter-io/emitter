package broker

import (
	"testing"

	"github.com/emitter-io/emitter/broker/message"
	netmock "github.com/emitter-io/emitter/network/mock"
	"github.com/emitter-io/emitter/security"
	secmock "github.com/emitter-io/emitter/security/mock"
	"github.com/emitter-io/emitter/security/usage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_onMe(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)
	s := &Service{
		subscriptions: message.NewTrie(),
		License:       license,
	}

	conn := netmock.NewConn()
	nc := s.newConn(conn.Client)
	resp, success := nc.onMe()
	meResp := resp.(*meResponse)

	assert.Equal(t, success, true, success)
	assert.NotNil(t, resp)
	assert.NotZero(t, len(meResp.ID))
}

func TestHandlers_onSubscribeUnsubscribe(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)
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
			subErr:        (*EventError)(nil),
			unsubCount:    0,
			unsubErr:      (*EventError)(nil),
			contractValid: true,
			contractFound: true,
			msg:           "Successful case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a+q/b/c/",
			subCount:      0,
			subErr:        ErrBadRequest,
			unsubCount:    0,
			unsubErr:      ErrBadRequest,
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
			subErr:        ErrUnauthorized,
			unsubCount:    0,
			unsubErr:      ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Expired key case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			subCount:      0,
			subErr:        ErrNotFound,
			unsubCount:    0,
			unsubErr:      ErrNotFound,
			contractValid: true,
			contractFound: false,
			msg:           "Contract not found case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			subCount:      0,
			subErr:        ErrUnauthorized,
			unsubCount:    0,
			unsubErr:      ErrUnauthorized,
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
			subErr:        ErrUnauthorized,
			unsubCount:    0,
			unsubErr:      ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Wrong target case",
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
			presence:      make(chan *presenceNotify, 100),
		}

		conn := netmock.NewConn()
		nc := s.newConn(conn.Client)
		s.Cipher, _ = s.License.Cipher()

		// Subscribe and check for error.
		subErr := nc.onSubscribe([]byte(tc.channel))
		assert.Equal(t, tc.subErr, subErr, tc.msg)

		// Search for the ssid.
		channel := security.ParseChannel([]byte(tc.channel))
		key, _ := s.Cipher.DecryptKey(channel.Key)
		ssid := message.NewSsid(key.Contract(), channel)
		subscribers := s.subscriptions.Lookup(ssid)
		assert.Equal(t, tc.subCount, len(subscribers))

		// Unsubscribe and check for error.
		unsubErr := nc.onUnsubscribe([]byte(tc.channel))
		assert.Equal(t, tc.unsubErr, unsubErr, tc.msg)

		// Search for the ssid.
		subscribers = s.subscriptions.Lookup(ssid)
		assert.Equal(t, tc.unsubCount, len(subscribers))

	}
}

func TestHandlers_onPublish(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)
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
			err:           (*EventError)(nil),
			contractValid: true,
			contractFound: true,
			msg:           "Successful case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a+q/b/c/",
			payload:       "test",
			err:           ErrBadRequest,
			contractValid: true,
			contractFound: true,
			msg:           "Invalid channel case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/+/b/c/",
			payload:       "test",
			err:           ErrForbidden,
			contractValid: true,
			contractFound: true,
			msg:           "Channel is not static case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBZRqJDby30mT/a/b/c/",
			payload:       "test",
			err:           ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Expired key case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			payload:       "test",
			err:           ErrNotFound,
			contractValid: true,
			contractFound: false,
			msg:           "Contract not found case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			payload:       "test",
			err:           ErrUnauthorized,
			contractValid: false,
			contractFound: true,
			msg:           "Contract is invalid case",
		},
		{
			channel:       "0Nq8SWbL8qoJzie4_C4yvupug6cLLlWO/a/b/c/",
			payload:       "test",
			err:           ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "No write permission case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBZHmCtcvoHGQ/a/b/c/",
			payload:       "test",
			err:           ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Wrong target case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			payload:       "test",
			err:           (*EventError)(nil),
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

		s := &Service{
			contracts:     provider,
			subscriptions: message.NewTrie(),
			License:       license,
		}

		conn := netmock.NewConn()
		nc := s.newConn(conn.Client)
		s.Cipher, _ = s.License.Cipher()

		err := nc.onPublish([]byte(tc.channel), []byte(tc.payload))

		assert.Equal(t, tc.err, err, tc.msg)
	}
}

func TestHandlers_onPresence(t *testing.T) {
	// TODO :
	// - valid key for the right channel, but no presence right.
	// - test Who
	license, _ := security.ParseLicense(testLicense)
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
			payload:       "{\"key\":\"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY\",\"channel\":\"a\",\"status\":true}",
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
			err:           ErrBadRequest,
			success:       false,
			contractValid: true,
			contractFound: true,
			msg:           "Invalid payload case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY\",\"channel\":\"a+b\",\"status\":true}",
			contractValid: true,
			contractFound: true,
			success:       false,
			err:           ErrBadRequest,
			msg:           "Invalid channel case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"0Nq8SWbL8qoOKEDqh_ebBZRqJDby30m\",\"channel\":\"a\",\"status\":true}",
			contractValid: true,
			contractFound: true,
			success:       false,
			err:           ErrUnauthorized,
			msg:           "Key for wrong channel case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY\",\"channel\":\"a+b\",\"status\":true}",
			err:           ErrNotFound,
			contractValid: true,
			contractFound: false,
			msg:           "Contract not found case",
		},
		{
			channel:       "emitter/presence/",
			payload:       "{\"key\":\"VfW_Cv5wWVZPHgCvLwJAuU2bgRFKXQEY\",\"channel\":\"a+b\",\"status\":true}",
			err:           ErrUnauthorized,
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

		conn := netmock.NewConn()
		nc := s.newConn(conn.Client)
		s.Cipher, _ = s.License.Cipher()

		resp, success := nc.onPresence([]byte(tc.payload))

		assert.Equal(t, tc.success, success, tc.msg)
		if !success {
			assert.Equal(t, tc.err, resp, tc.msg)
		}
	}
}

func TestHandlers_onKeygen(t *testing.T) {
	license, _ := security.ParseLicense("N7XxQbUEPxJ_RIj4muLUdLGYtR1kdKe2AAAAAAAAAAI")
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
			resp:          ErrBadRequest,
			msg:           "Invalid request case",
		},
		{
			payload:       "{\"key\":\"xEbaDPaICEwVhgdnl2rg_1DWi_MAg_3B\",\"channel\":\"article1\"}",
			contractValid: true,
			contractFound: true,
			generated:     false,
			resp:          ErrUnauthorized,
			msg:           "No keygen permission case (not a master key)",
		},
		{
			payload:       "{\"key\":\"8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR\",\"channel\":\"article1\"}",
			contractValid: true,
			contractFound: false,
			generated:     false,
			resp:          ErrNotFound,
			msg:           "Contract not found case",
		},
		{
			payload:       "{\"key\":\"8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR\",\"channel\":\"article1\"}",
			contractValid: false,
			contractFound: true,
			generated:     false,
			resp:          ErrUnauthorized,
			msg:           "Contract not valid case",
		},
		{
			payload:       "{\"key\":\"8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR\",\"channel\":\"article1\"}",
			contractValid: true,
			contractFound: true,
			generated:     true,
			resp:          keyGenResponse{Status: 200, Key: "76w5HdpyIOQh70HnB4d33gbqD5fFztGY", Channel: "article1"},
			msg:           "Successful case",
		},
	}

	//keyGenResponse{Status: 200, Key: "76w5HdpyIOQh70HnB4d33gbqD5fFztGY", Channel: "article1"},
	for _, tc := range tests {
		provider := secmock.NewContractProvider()
		contract := new(secmock.Contract)
		contract.On("Validate", mock.Anything).Return(tc.contractValid)
		contract.On("Stats").Return(usage.NewMeter(0))

		provider.On("Get", mock.Anything).Return(contract, tc.contractFound)
		/*
			if tc.contractFound {
				provider.On("Get", mock.Anything).Return(contract).Once()
			} else {
				provider.On("Get", mock.Anything).Return(nil).Once()
			}*/

		s := &Service{
			contracts:     provider,
			subscriptions: message.NewTrie(),
			License:       license,
		}

		conn := netmock.NewConn()
		nc := s.newConn(conn.Client)
		s.Cipher, _ = s.License.Cipher()

		//resp
		resp, generated := nc.onKeyGen([]byte(tc.payload))
		assert.Equal(t, tc.generated, generated, tc.msg)

		if !generated {
			keyGenResp := resp.(*EventError)
			assert.Equal(t, tc.resp, keyGenResp)
		} else {
			keyGenResp := resp.(*keyGenResponse)
			expected := tc.resp.(keyGenResponse)
			//assert.Equal(t, tc.resp, keyGenResp)
			assert.Equal(t, expected.Status, keyGenResp.Status)

		}
	}
}
