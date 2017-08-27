/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
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

package security

import (
	"errors"
	"fmt"
	"sync"

	"github.com/emitter-io/emitter/network/http"
)

// The contract's state possible values.
const (
	ContractStateAllowed = uint8(iota)
	ContractStateRefused
)

// Contract represents an interface for a contract.
type Contract interface {
	Validate(key Key) bool // Validate checks the security key with the contract.
	Stats() UsageStats     // Gets the usage statistics.
}

// contract represents a contract (user account).
type contract struct {
	ID        uint32 `json:"id"`     // Gets or sets the contract id.
	MasterID  uint16 `json:"sign"`   // Gets or sets the master id.
	Signature uint32 `json:"master"` // Gets or sets the signature of the contract.
	State     uint8  `json:"state"`  // Gets or sets the state of the contract.
	stats     *usage // Gets the usage stats.
}

// Validate validates the contract data against a key.
func (c *contract) Validate(key Key) bool {
	return c.MasterID == key.Master() &&
		c.Signature == key.Signature() &&
		c.ID == key.Contract() &&
		c.State == ContractStateAllowed
}

// Gets the usage statistics.
func (c *contract) Stats() UsageStats {
	return c.stats
}

// ContractProvider represents an interface for a contract provider.
type ContractProvider interface {
	// Creates a new instance of a Contract in the underlying data storage.
	Create() (Contract, error)
	Get(id uint32) Contract
}

// SingleContractProvider provides contracts on premise.
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
	p.owner.stats = new(usage)
	return p
}

// Create creates a contract, the SingleContractProvider way.
func (p *SingleContractProvider) Create() (Contract, error) {
	return nil, errors.New("Single contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *SingleContractProvider) Get(id uint32) Contract {
	if p.owner == nil || p.owner.ID != id {
		return nil
	}
	return p.owner
}

// HTTPContractProvider provides contracts over http.
type HTTPContractProvider struct {
	owner *contract
	cache *sync.Map
	//cache *collection.ConcurrentMap
}

// NewHTTPContractProvider creates a new single contract provider.
func NewHTTPContractProvider(license *License) *HTTPContractProvider {
	p := HTTPContractProvider{}
	p.owner = new(contract)
	p.owner.MasterID = 1
	p.owner.ID = license.Contract
	p.owner.Signature = license.Signature
	p.cache = new(sync.Map)

	return &p
}

// Create creates a contract, the HTTPContractProvider way.
func (p *HTTPContractProvider) Create() (Contract, error) {
	return nil, errors.New("HTTP contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *HTTPContractProvider) Get(id uint32) Contract {
	if c, ok := p.cache.Load(id); ok {
		return c.(Contract)
	}

	// Load or store again, since we might have concurrently update it meanwhile
	contract := p.fetchContract(id)
	c, _ := p.cache.LoadOrStore(id, contract)
	return c.(Contract)
}

func (p *HTTPContractProvider) fetchContract(id uint32) *contract {
	c := &contract{
		stats: new(usage),
	}

	query := fmt.Sprintf("http://meta.emitter.io/v1/contract/%d", int32(id)) // meta currently requires a signed int
	err := http.Get(query, c)

	if err != nil || c.ID == 0 {
		return nil
	}

	return c
}
