package security

import (
	"errors"
)

// Contract represents an interface for a contract.
type Contract interface {
	Validate(key Key) bool // Validate checks the security key with the contract.
}

// contract represents a contract (user account).
type contract struct {
	ID        int32  // Gets or sets the contract id.
	MasterID  uint16 // Gets or sets the master id.
	Signature int32  // Gets or sets the signature of the contract.
}

// Validate validates the contract data against a key.
func (c *contract) Validate(key Key) bool {
	return c.MasterID == key.Master() &&
		c.Signature == key.Signature() &&
		c.ID == key.Contract()
}

// ContractProvider represents an interface for a contract provider.
type ContractProvider interface {
	// Creates a new instance of a Contract in the underlying data storage.
	Create() Contract
	Get(id int32) Contract
}

//SingleContractProvider provide contracts on premise.
type SingleContractProvider struct {
	owner *contract
}

// NewSingleContractProvider creates a new single contract provider.
func NewSingleContractProvider(license *License) *SingleContractProvider {
	p := new(SingleContractProvider)
	p.owner = new(contract)
	p.owner.MasterID = 1
	p.owner.ID = license.Contract
	p.owner.Signature = license.Signature
	return p
}

// Create a contract, the SingleContractProvider way.
func (p *SingleContractProvider) Create(license *License) (Contract, error) {
	return nil, errors.New("Single contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *SingleContractProvider) Get(id int32) Contract {
	if p.owner == nil || p.owner.ID != id {
		return nil
	}
	return p.owner
}
