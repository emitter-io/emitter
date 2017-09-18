package mock

import (
	"testing"

	"github.com/emitter-io/emitter/security/usage"
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
