package mock

import (
	"errors"
	"github.com/emitter-io/emitter/security"
)

// contract represents a contract (user account).
type contract struct {
	ID        uint32 `json:"id"`     // Gets or sets the contract id.
	MasterID  uint16 `json:"sign"`   // Gets or sets the master id.
	Signature uint32 `json:"master"` // Gets or sets the signature of the contract.
	State     uint8  `json:"state"`
}

// Validate validates the contract data against a key.
func (c *contract) Validate(key security.Key) bool {
	return c.MasterID == key.Master() &&
		c.Signature == key.Signature() &&
		c.ID == key.Contract() &&
		c.State == security.ContractStateAllowed
}

// InvalidContractProvider provides contracts on premise.
type InvalidContractProvider struct {
}

// NewInvalidContractProvider creates a new single contract provider that will only provide invalid contracts.
func NewInvalidContractProvider(license *security.License) *InvalidContractProvider {
	p := new(InvalidContractProvider)
	return p
}

// Create creates a contract.
func (p *InvalidContractProvider) Create() (security.Contract, error) {
	return nil, errors.New("Single contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *InvalidContractProvider) Get(id uint32) security.Contract {
	return &contract{}
}

// NotFoundContractProvider won't find any contracts.
type NotFoundContractProvider struct {
}

// NewNotfoundContractProvider creates a new provider that won't find any contracts.
func NewNotFoundContractProvider(license *security.License) *InvalidContractProvider {
	p := new(InvalidContractProvider)
	return p
}

// Create creates a contract.
func (p *NotFoundContractProvider) Create() (security.Contract, error) {
	return nil, errors.New("Single contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *NotFoundContractProvider) Get(id uint32) security.Contract {
	return nil
}
