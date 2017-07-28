package security

//package mock

import (
	"errors"
	//"github.com/emitter-io/emitter/security"
)

// InvalidContractProvider provides contracts on premise.
type InvalidContractProvider struct {
}

// NewInvalidContractProvider creates a new single contract provider that will only provide invalid contracts.
func NewInvalidContractProvider(license *License) *InvalidContractProvider {
	p := new(InvalidContractProvider)
	return p
}

// Create creates a contract.
func (p *InvalidContractProvider) Create() (Contract, error) {
	return nil, errors.New("Single contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *InvalidContractProvider) Get(id uint32) Contract {
	return &contract{}
}

// NotFoundContractProvider won't find any contracts.
type NotFoundContractProvider struct {
}

// NewNotfoundContractProvider creates a new provider that won't find any contracts.
func NewNotFoundContractProvider(license *License) *InvalidContractProvider {
	p := new(InvalidContractProvider)
	return p
}

// Create creates a contract.
func (p *NotFoundContractProvider) Create() (Contract, error) {
	return nil, errors.New("Single contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *NotFoundContractProvider) Get(id uint32) Contract {
	return nil
}
