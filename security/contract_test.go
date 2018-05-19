package security

import (
	"encoding/json"
	"testing"

	"github.com/emitter-io/emitter/network/http"
	"github.com/emitter-io/emitter/security/usage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func testNewSingleContractProvider() (*SingleContractProvider, *License) {
	license, _ := ParseLicense("zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI")
	return NewSingleContractProvider(license, new(usage.NoopStorage)), license
}

func testNewHTTPContractProvider() (*HTTPContractProvider, *License) {
	license, _ := ParseLicense("zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI")
	return NewHTTPContractProvider(license, new(usage.NoopStorage)), license
}

func TestSingleContractProvider_Name(t *testing.T) {
	p := SingleContractProvider{}
	assert.Equal(t, "single", p.Name())
}

func TestNewSingleContractProvider(t *testing.T) {
	p, license := testNewSingleContractProvider()

	assert.EqualValues(t, p.owner.MasterID, 1)
	assert.EqualValues(t, p.owner.Signature, license.Signature)
	assert.EqualValues(t, p.owner.ID, license.Contract)
	assert.NotNil(t, p.owner.Stats())
}

func TestSingleContractProvider_Create(t *testing.T) {
	p, _ := testNewSingleContractProvider()

	err := p.Configure(nil)
	assert.NoError(t, err)

	contract, err := p.Create()
	assert.Nil(t, contract)
	assert.Error(t, err)
}

func TestSingleContractProvider_Get(t *testing.T) {
	p, license := testNewSingleContractProvider()
	contractByID, ok1 := p.Get(license.Contract)
	assert.True(t, ok1)
	assert.NotNil(t, contractByID)

	contractByWrongID, ok2 := p.Get(0)
	assert.False(t, ok2)
	assert.Nil(t, contractByWrongID)
}

func TestSingleContractProvider_Validate(t *testing.T) {
	p, license := testNewSingleContractProvider()
	contract, ok := p.Get(license.Contract)
	assert.True(t, ok)

	key := Key(make([]byte, 24))
	key.SetMaster(1)
	key.SetContract(license.Contract)
	key.SetSignature(license.Signature)

	assert.True(t, contract.Validate(key))
}

func TestNewHTTPContractProvider(t *testing.T) {
	p, license := testNewHTTPContractProvider()

	assert.EqualValues(t, p.owner.MasterID, 1)
	assert.EqualValues(t, p.owner.Signature, license.Signature)
	assert.EqualValues(t, p.owner.ID, license.Contract)
}

func TestHTTPContractProvider_Create(t *testing.T) {
	p, _ := testNewHTTPContractProvider()

	contract, err := p.Create()

	assert.Nil(t, contract)
	assert.Error(t, err)
}

func TestHTTPContractProvider_Name(t *testing.T) {
	p := HTTPContractProvider{}
	assert.Equal(t, "http", p.Name())
}

func TestHTTPContractProvider_Get(t *testing.T) {

	h := http.NewMockClient()
	h.On("Get", "1", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		output := args.Get(1).(interface{})
		json.Unmarshal([]byte(`{"id": 1}`), output)
	}).Return([]byte{}, nil)

	h.On("Get", "2", mock.Anything, mock.Anything).Return([]byte{}, nil)

	p, _ := testNewHTTPContractProvider()
	p.http = h

	contractByID, ok1 := p.Get(1)
	assert.True(t, ok1)
	assert.NotNil(t, contractByID)

	contractByWrongID, ok2 := p.Get(2)
	assert.False(t, ok2)
	assert.Nil(t, contractByWrongID)
}

func TestHTTPContractPovider_Configure(t *testing.T) {
	p, _ := testNewHTTPContractProvider()

	{
		err := p.Configure(nil)
		assert.Error(t, err)
	}

	{
		err := p.Configure(map[string]interface{}{})
		assert.Error(t, err)
	}

	{
		err := p.Configure(map[string]interface{}{
			"authorization": "Digest 123",
			"url":           "http://127.0.0.1",
			"interval":      600000.0,
		})
		assert.NoError(t, err)
		assert.NoError(t, p.Close())
	}
}

func TestHTTPContractPovider_refresh(t *testing.T) {
	h := http.NewMockClient()
	h.On("Get", "1", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		output := args.Get(1).(interface{})
		json.Unmarshal([]byte(`{"id": 1, "state": 2}`), output)
	}).Return([]byte{}, nil)

	p, _ := testNewHTTPContractProvider()
	p.http = h

	p.cache.Store(uint32(1), nil)
	p.refresh()

	c, ok := p.cache.Load(uint32(1))
	assert.True(t, ok)
	assert.NotNil(t, c)
	assert.Equal(t, uint8(2), c.(*contract).State)
}

func TestNoopContractPovider(t *testing.T) {
	p := NewNoopContractProvider()

	err := p.Configure(nil)
	assert.NoError(t, err)

	c, err := p.Create()
	assert.Nil(t, c)
	assert.Equal(t, false, err == nil)

	_, ok := p.Get(10)
	assert.False(t, ok)

	n := p.Name()
	assert.Equal(t, "noop", n)
}
