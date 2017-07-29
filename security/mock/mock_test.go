package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMock(t *testing.T) {
	c := new(Contract)
	c.On("Validate", mock.Anything).Return(true)

	m := NewContractProvider()
	id := uint32(1)

	m.On("Get", id).Return(c)
	assert.Equal(t, c, m.Get(id))
	assert.True(t, m.Get(id).Validate(nil), true)

	m.On("Create").Return(c, nil)
	contract, err := m.Create()
	assert.Equal(t, c, contract)
	assert.NoError(t, err)
}
