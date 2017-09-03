package broker

import (
	"testing"

	"github.com/emitter-io/emitter/broker/subscription"
	netmock "github.com/emitter-io/emitter/network/mock"
	"github.com/emitter-io/emitter/security"
	secmock "github.com/emitter-io/emitter/security/mock"
	"github.com/emitter-io/emitter/security/usage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
		if tc.contractFound {
			provider.On("Get", mock.Anything).Return(contract)
		} else {
			provider.On("Get", mock.Anything).Return(nil)
		}

		s := &Service{
			contracts:     provider,
			subscriptions: subscription.NewTrie(),
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
		ssid := subscription.NewSsid(key.Contract(), channel)
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
		if tc.contractFound {
			provider.On("Get", mock.Anything).Return(contract).Once()
		} else {
			provider.On("Get", mock.Anything).Return(nil).Once()
		}

		s := &Service{
			contracts:     provider,
			subscriptions: subscription.NewTrie(),
			License:       license,
		}

		conn := netmock.NewConn()
		nc := s.newConn(conn.Client)
		s.Cipher, _ = s.License.Cipher()

		err := nc.onPublish([]byte(tc.channel), []byte(tc.payload))

		assert.Equal(t, tc.err, err, tc.msg)
	}
}

/*
func TestHandlers_onKeygen(t *testing.T) {
	license, _ := security.ParseLicense("pLcaYvemMQOZR9o9sa5COWztxfAAAAAAAAAAAAAAAAI")
	tests := []struct {
		payload string
		//err           error
		contractValid bool
		contractFound bool
		generated     bool
		resp          keyGenResponse
		msg           string
	}{
		{
			payload: "{\"key\":\"xEbaDPaICEwVhgdnl2rg_1DWi_MAg_3B\",\"channel\":\"article1\"}",
			//err:           (*EventError)(nil),
			contractValid: true,
			contractFound: true,
			generated:     true,
			resp:          keyGenResponse{Status: 200, Key: "76w5HdpyIOQh70HnB4d33gbqD5fFztGY", Channel: "article1"},
			msg:           "Successful case",
		}, /*
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
			},*/ /*
	}

	for _, tc := range tests {
		//HOW COULD I MOCK KEY SET SALT??????????????????????????????????????????

		security.Key.SetSalt = (k Key) func (value uint16) {
			k[0] = byte(value >> 8)
			k[1] = byte(value)
		}


		contract := new(secmock.Contract)
		contract.On("Validate", mock.Anything).Return(tc.contractValid)
		contract.On("Stats").Return(security.NewUsageStats())

		provider := secmock.NewContractProvider()
		if tc.contractFound {
			provider.On("Get", mock.Anything).Return(contract).Once()
		} else {
			provider.On("Get", mock.Anything).Return(nil).Once()
		}

		s := &Service{
			Contracts:     provider,
			subscriptions: subscription.NewTrie(),
			License:       license,
		}

		conn := netmock.NewConn()
		nc := s.newConn(conn.Client)
		s.Cipher, _ = s.License.Cipher()

		//resp
		resp, generated := nc.onKeyGen([]byte(tc.payload))
		assert.Equal(t, tc.generated, generated, tc.msg)

		if generated {
			keyGenResp := resp.(*keyGenResponse)
			assert.Equal(t, tc.resp, keyGenResp)

		}
	}
}*/
