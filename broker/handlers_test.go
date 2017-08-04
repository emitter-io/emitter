package broker

import (
	"testing"

	netmock "github.com/emitter-io/emitter/network/mock"
	"github.com/emitter-io/emitter/security"
	secmock "github.com/emitter-io/emitter/security/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_onSubscribe(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)
	tests := []struct {
		channel       string
		count         int
		err           error
		contractValid bool
		contractFound bool
		msg           string
	}{
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			count:         1,
			err:           (*EventError)(nil),
			contractValid: true,
			contractFound: true,
			msg:           "Successful case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a+q/b/c/",
			count:         0,
			err:           ErrBadRequest,
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
			count:         0,
			err:           ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Expired key case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			count:         0,
			err:           ErrNotFound,
			contractValid: true,
			contractFound: false,
			msg:           "Contract not found case",
		},
		{
			channel:       "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/",
			count:         0,
			err:           ErrUnauthorized,
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
			count:         0,
			err:           ErrUnauthorized,
			contractValid: true,
			contractFound: true,
			msg:           "Wrong target case",
		},
	}

	for _, tc := range tests {

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
		Contracts:     security.NewSingleContractProvider(license),
		subscriptions: NewSubscriptionTrie(),
		License:       license,
		subcounters:   NewSubscriptionCounters(),
		}

		conn := netmock.NewConn()
		nc := s.newConn(conn.Client)
		s.Cipher, _ = s.License.Cipher()

		err := nc.onSubscribe([]byte(tc.channel))

		assert.Equal(t, tc.err, err, tc.msg)
		c := s.subcounters.All()
		assert.Equal(t, tc.count, len(c))
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
		contract.On("Stats").Return(security.NewUsageStats())

		provider := secmock.NewContractProvider()
		if tc.contractFound {
			provider.On("Get", mock.Anything).Return(contract).Once()
		} else {
			provider.On("Get", mock.Anything).Return(nil).Once()
		}

		s := &Service{
			Contracts:     provider,
			subscriptions: NewSubscriptionTrie(),
			License:       license,
			subcounters:   NewSubscriptionCounters(),
		}

		conn := netmock.NewConn()
		nc := s.newConn(conn.Client)
		s.Cipher, _ = s.License.Cipher()

		err := nc.onPublish([]byte(tc.channel), []byte(tc.payload))

		assert.Equal(t, tc.err, err, tc.msg)
	}
}

func TestHandlers_onUnsubscribe(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)
	s := &Service{
		Contracts:     security.NewSingleContractProvider(license),
		subscriptions: NewSubscriptionTrie(),
		License:       license,
		subcounters:   NewSubscriptionCounters(),
	}

	conn := netmock.NewConn()
	nc := s.newConn(conn.Client)
	s.Cipher, _ = s.License.Cipher()
	err := nc.onUnsubscribe([]byte("0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/"))

	assert.Nil(t, err)
	//assert.Equal(t, 1, s.Counters.GetCounter().Value())
}
