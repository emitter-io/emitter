/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
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

package contract

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/emitter-io/config"
	"github.com/emitter-io/emitter/internal/async"
	"github.com/emitter-io/emitter/internal/network/http"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/security/license"
)

// The contract's state possible values.
const (
	ContractStateUnknown = uint8(iota)
	ContractStateAllowed
	ContractStateRefused
)

// Contract represents an interface for a contract.
type Contract interface {
	Validate(key security.Key) bool // Validate checks the security key with the contract.
	Stats() usage.Meter             // Gets the usage statistics.
}

// contract represents a contract (user account).
type contract struct {
	ID        uint32      `json:"id"`     // Gets or sets the contract id.
	MasterID  uint16      `json:"master"` // Gets or sets the master id.
	Signature uint32      `json:"sign"`   // Gets or sets the signature of the contract.
	State     uint8       `json:"state"`  // Gets or sets the state of the contract.
	stats     usage.Meter // Gets the usage stats.
}

// Validate validates the contract data against a key.
func (c *contract) Validate(key security.Key) bool {
	return c.MasterID == key.Master() &&
		c.Signature == key.Signature() &&
		c.ID == key.Contract() &&
		c.State == ContractStateAllowed
}

// Stats gets the usage statistics.
func (c *contract) Stats() usage.Meter {
	return c.stats
}

// Provider represents an interface for a contract provider.
type Provider interface {
	config.Provider

	Create() (Contract, error)
	Get(id uint32) (Contract, bool)
}

// ------------------------------------------------------------------------------------

// Assert interface compliance
var _ Provider = new(HTTPContractProvider)

// NoopContractProvider does not provide a contract.
type NoopContractProvider struct{}

// NewNoopContractProvider creates a new no-op contract provider.
func NewNoopContractProvider() *NoopContractProvider {
	return new(NoopContractProvider)
}

// Name returns the name of the provider.
func (p *NoopContractProvider) Name() string {
	return "noop"
}

// Configure configures the provider.
func (p *NoopContractProvider) Configure(config map[string]interface{}) error {
	return nil
}

// Create creates a contract, the SingleContractProvider way.
func (p *NoopContractProvider) Create() (Contract, error) {
	return nil, errors.New("Noop contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *NoopContractProvider) Get(id uint32) (Contract, bool) {
	return nil, false
}

// ------------------------------------------------------------------------------------

// Assert interface compliance
var _ Provider = new(SingleContractProvider)

// SingleContractProvider provides contracts on premise.
type SingleContractProvider struct {
	owner *contract      // The owner contract.
	usage usage.Metering // The usage stats container.
}

// NewSingleContractProvider creates a new single contract provider.
func NewSingleContractProvider(license license.License, metering usage.Metering) *SingleContractProvider {
	p := new(SingleContractProvider)
	p.owner = new(contract)
	p.owner.MasterID = 1
	p.owner.ID = license.Contract()
	p.owner.Signature = license.Signature()
	p.owner.State = ContractStateAllowed
	p.usage = metering
	p.owner.stats = p.usage.Get(license.Contract()).(usage.Meter)
	return p
}

// Name returns the name of the provider.
func (p *SingleContractProvider) Name() string {
	return "single"
}

// Configure configures the provider.
func (p *SingleContractProvider) Configure(config map[string]interface{}) error {
	return nil
}

// Create creates a contract, the SingleContractProvider way.
func (p *SingleContractProvider) Create() (Contract, error) {
	return nil, errors.New("Single contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *SingleContractProvider) Get(id uint32) (Contract, bool) {
	if p.owner == nil || p.owner.ID != id {
		return nil, false
	}

	return p.owner, true
}

// ------------------------------------------------------------------------------------

// Assert interface compliance
var _ Provider = new(HTTPContractProvider)

// HTTPContractProvider provides contracts over http.
type HTTPContractProvider struct {
	url    string             // The url to hit for the provider.
	owner  *contract          // The owner contract.
	cache  *sync.Map          // The cache for the contracts.
	usage  usage.Metering     // The usage stats container.
	http   http.Client        // The http client to use.
	head   []http.HeaderValue // The http headers to add with each request.
	cancel context.CancelFunc // The cancellation function.
}

// NewHTTPContractProvider creates a new single contract provider.
func NewHTTPContractProvider(license license.License, metering usage.Metering) *HTTPContractProvider {
	p := HTTPContractProvider{}
	p.owner = new(contract)
	p.owner.MasterID = 1
	p.owner.ID = license.Contract()
	p.owner.Signature = license.Signature()
	p.cache = new(sync.Map)
	p.usage = metering
	return &p
}

// Name returns the name of the provider.
func (p *HTTPContractProvider) Name() string {
	return "http"
}

// Configure configures the provider.
func (p *HTTPContractProvider) Configure(config map[string]interface{}) (err error) {
	if config == nil {
		return errors.New("Configuration was not provided for HTTP contract provider")
	}

	// Get the interval from the provider configuration
	interval := 10 * time.Minute
	if v, ok := config["interval"]; ok {
		if i, ok := v.(float64); ok {
			interval = time.Duration(i) * time.Millisecond
		}
	}

	// Get the authorization header to add to the request
	headers := []http.HeaderValue{http.NewHeader("Accept", "application/json")}
	if v, ok := config["authorization"]; ok {
		if header, ok := v.(string); ok {
			headers = append(headers, http.NewHeader("Authorization", header))
		}
	}

	// Get the url from the provider configuration
	if url, ok := config["url"]; ok {
		p.url = url.(string)

		// Create a new HTTP client to use
		p.http, err = http.NewClient(10 * time.Second)
		p.head = headers

		// Periodically refresh contracts
		p.cancel = async.Repeat(context.Background(), interval, p.refresh)
		return
	}

	return errors.New("The 'url' parameter was not provider in the configuration for HTTP contract provider")
}

// Create creates a contract, the HTTPContractProvider way.
func (p *HTTPContractProvider) Create() (Contract, error) {
	return nil, errors.New("HTTP contract provider can not create contracts")
}

// Get returns a ContractData fetched by its id.
func (p *HTTPContractProvider) Get(id uint32) (Contract, bool) {
	if c, ok := p.cache.Load(id); ok {
		return c.(Contract), true
	}

	// Load or store again, since we might have concurrently update it meanwhile
	if contract, ok := p.fetchContract(id); ok {
		c, _ := p.cache.LoadOrStore(id, contract)
		return c.(Contract), true
	}

	return nil, false
}

// Close closes the provider.
func (p *HTTPContractProvider) Close() error {
	if p.cancel != nil {
		p.cancel()
	}

	return nil
}

// Fetches a single contract from the underlying contract provider.
func (p *HTTPContractProvider) fetchContract(id uint32) (*contract, bool) {
	c := new(contract)
	_, err := p.http.Get(fmt.Sprintf("%s%d", p.url, id), c, p.head...)
	if err != nil {
		logging.LogError("contract", "fetching http contract", err)
		return nil, false
	}

	if c.ID == 0 {
		return nil, false
	}

	c.stats = p.usage.Get(id).(usage.Meter)
	return c, true
}

// Refresh fetches all the contracts from the underlying contract provider.
func (p *HTTPContractProvider) refresh() {
	p.cache.Range(func(k, v interface{}) bool {
		if id, ok := k.(uint32); ok {
			if contract, ok := p.fetchContract(id); ok {
				p.cache.Store(id, contract)
			}
		}
		return true
	})
}
