package mock

import (
	"github.com/emitter-io/emitter/security"
	"github.com/stretchr/testify/mock"
)

// Contract represents a contract (user account).
type Contract struct {
	mock.Mock
}

// Validate validates the contract data against a key.
func (mock *Contract) Validate(key security.Key) bool {
	mockArgs := mock.Called(key)
	return mockArgs.Get(0).(bool)
}

// ContractProvider is the mock provider for contracts
type ContractProvider struct {
	mock.Mock
}

// NewContractProvider creates a new mock client provider.
func NewContractProvider() *ContractProvider {
	return new(ContractProvider)
}

// Create creates a contract.
func (mock *ContractProvider) Create() (security.Contract, error) {
	mockArgs := mock.Called()
	return mockArgs.Get(0).(security.Contract), mockArgs.Error(1)
}

// Get returns a ContractData fetched by its id.
func (mock *ContractProvider) Get(id uint32) security.Contract {
	mockArgs := mock.Called(id)
	return mockArgs.Get(0).(security.Contract)
}
