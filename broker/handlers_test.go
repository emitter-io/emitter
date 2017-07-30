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
	s := &Service{
		ContractProvider: security.NewSingleContractProvider(license),
		subscriptions:    NewSubscriptionTrie(),
		License:          license,
		subcounters:      NewSubscriptionCounters(),
	}

	conn := netmock.NewConn()
	nc := s.newConn(conn.Client)
	s.Cipher, _ = s.License.Cipher()
	err := nc.onSubscribe([]byte("0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/"))

	assert.Nil(t, err)
	//assert.Equal(t, 1, s.Counters.GetCounter().Value())
}

func TestHandlers_onPublish(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)

	validContract := new(secmock.Contract)
	invalidContract := new(secmock.Contract)
	validContract.On("Validate", mock.Anything).Return(true)
	invalidContract.On("Validate", mock.Anything).Return(false)

	singleContractProvider := secmock.NewContractProvider()
	singleContractProvider.On("Get", mock.Anything).Return(validContract)
	singleContractProvider.On("Create").Return(validContract, nil)

	invalidContractProvider := secmock.NewContractProvider()
	invalidContractProvider.On("Get", mock.Anything).Return(invalidContract)
	invalidContractProvider.On("Create").Return(invalidContract, nil)

	notFoundContractProvider := secmock.NewContractProvider()
	notFoundContractProvider.On("Get", mock.Anything).Return(nil)
	notFoundContractProvider.On("Create").Return(nil, nil)

	s := &Service{
		ContractProvider: singleContractProvider,
		subscriptions:    NewSubscriptionTrie(),
		License:          license,
		subcounters:      NewSubscriptionCounters(),
	}

	conn := netmock.NewConn()
	nc := s.newConn(conn.Client)
	s.Cipher, _ = s.License.Cipher()

	tests := []struct {
		channel          string
		payload          string
		err              error
		contractProvider security.ContractProvider
	}{
		// Successful.
		{channel: "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/", payload: "test", err: (*EventError)(nil), contractProvider: singleContractProvider},

		// Channel is invalid.
		{channel: "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a+q/b/c/", payload: "test", err: ErrBadRequest, contractProvider: singleContractProvider},

		// Channel is not static.
		{channel: "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/+/b/c/", payload: "test", err: ErrForbidden, contractProvider: singleContractProvider},

		// The key could not be decrypted.

		// Key is expired.
		{channel: "0Nq8SWbL8qoOKEDqh_ebBZRqJDby30mT/a/b/c/", payload: "test", err: ErrUnauthorized, contractProvider: singleContractProvider},

		// Contract not found.
		{channel: "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/", payload: "test", err: ErrNotFound, contractProvider: notFoundContractProvider},

		// Contract is invalid.
		{channel: "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/", payload: "test", err: ErrUnauthorized, contractProvider: invalidContractProvider},

		// Key does not provide the permission to write.
		{channel: "0Nq8SWbL8qoJzie4_C4yvupug6cLLlWO/a/b/c/", payload: "test", err: ErrUnauthorized, contractProvider: singleContractProvider},

		// Key does not provide the permission for that channel.
		{channel: "0Nq8SWbL8qoOKEDqh_ebBZHmCtcvoHGQ/a/b/c/", payload: "test", err: ErrUnauthorized, contractProvider: singleContractProvider},

		// A TTL is specified but the key does not provide the permission to store.
		{channel: "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/", payload: "test", err: (*EventError)(nil), contractProvider: singleContractProvider},
	}

	for _, tc := range tests {
		s.ContractProvider = tc.contractProvider
		err := nc.onPublish([]byte(tc.channel), []byte(tc.payload))

		assert.Equal(t, tc.err, err)

	}
}

func TestHandlers_onUnsubscribe(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)
	s := &Service{
		ContractProvider: security.NewSingleContractProvider(license),
		subscriptions:    NewSubscriptionTrie(),
		License:          license,
		subcounters:      NewSubscriptionCounters(),
	}

	conn := netmock.NewConn()
	nc := s.newConn(conn.Client)
	s.Cipher, _ = s.License.Cipher()
	err := nc.onUnsubscribe([]byte("0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO/a/b/c/"))

	assert.Nil(t, err)
	//assert.Equal(t, 1, s.Counters.GetCounter().Value())
}
