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
	"testing"

	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMock(t *testing.T) {

	id := uint32(1)

	c := new(Contract)
	c.On("Validate", mock.Anything).Return(true)
	c.On("Stats").Return(usage.NewMeter(id))

	m := NewContractProvider()
	cfg := make(map[string]interface{})

	assert.Equal(t, "mock", m.Name())

	m.On("Configure", cfg).Return(nil)
	assert.NoError(t, m.Configure(cfg))

	m.On("Get", id).Return(c, true)
	c1, o1 := m.Get(id)
	assert.True(t, o1)
	assert.Equal(t, c, c1)
	assert.True(t, c1.Validate(nil), true)
	assert.Equal(t, usage.NewMeter(id), c1.Stats())

	m.On("Create").Return(c, nil)
	contract, err := m.Create()
	assert.Equal(t, c, contract)
	assert.NoError(t, err)
}
