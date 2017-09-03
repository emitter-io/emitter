package mock

import (
	"github.com/emitter-io/emitter/security"
	"github.com/emitter-io/emitter/security/usage"
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

// Stats returns the stats.
func (mock *Contract) Stats() usage.Meter {
	mockArgs := mock.Called()
	return mockArgs.Get(0).(usage.Meter)
}

// ContractProvider is the mock provider for contracts
type ContractProvider struct {
	mock.Mock
}

// NewContractProvider creates a new mock client provider.
func NewContractProvider() *ContractProvider {
	return new(ContractProvider)
}

// Name returns the name of the provider.
func (mock *ContractProvider) Name() string {
	return "mock"
}

// Configure configures the provider.
func (mock *ContractProvider) Configure(config map[string]interface{}) error {
	mockArgs := mock.Called(config)
	return mockArgs.Error(0)
}

// Create creates a contract.
func (mock *ContractProvider) Create() (security.Contract, error) {
	mockArgs := mock.Called()
	return mockArgs.Get(0).(security.Contract), mockArgs.Error(1)
}

// Get returns a ContractData fetched by its id.
func (mock *ContractProvider) Get(id uint32) security.Contract {
	mockArgs := mock.Called(id)
	contract := mockArgs.Get(0)
	if contract != nil {
		return contract.(security.Contract)
	}
	return nil
}
