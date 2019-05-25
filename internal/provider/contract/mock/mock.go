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

package mock

import (
	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security"
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
func (mock *ContractProvider) Create() (contract.Contract, error) {
	mockArgs := mock.Called()
	return mockArgs.Get(0).(contract.Contract), mockArgs.Error(1)
}

// Get returns a ContractData fetched by its id.
func (mock *ContractProvider) Get(id uint32) (contract.Contract, bool) {
	mockArgs := mock.Called(id)
	return mockArgs.Get(0).(contract.Contract), mockArgs.Bool(1)
}
