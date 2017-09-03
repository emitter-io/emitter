package security

import (
	"encoding/json"
	"testing"

	"github.com/emitter-io/emitter/network/http"
	"github.com/emitter-io/emitter/security/usage"
	"github.com/stretchr/testify/assert"
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
	contract, err := p.Create()

	assert.Nil(t, contract)
	assert.Error(t, err)
}

func TestSingleContractProvider_Get(t *testing.T) {
	p, license := testNewSingleContractProvider()
	contractByID := p.Get(license.Contract)
	contractByWrongID := p.Get(0)

	assert.NotNil(t, contractByID)
	assert.Nil(t, contractByWrongID)
}

func TestSingleContractProvider_Validate(t *testing.T) {
	p, license := testNewSingleContractProvider()
	contract := p.Get(license.Contract)

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
	p, _ := testNewHTTPContractProvider()

	oldGet := http.Get
	defer func() {
		http.Get = oldGet
	}()

	http.Get = func(url string, output interface{}, headers ...http.HeaderValue) error {
		if url == "1" {
			return json.Unmarshal([]byte(`{"id": 1}`), output)
		}
		return nil
	}

	contractByID := p.Get(1)
	contractByID = p.Get(1)
	contractByWrongID := p.Get(0)
	http.Get = oldGet
	assert.NotNil(t, contractByID)
	assert.Nil(t, contractByWrongID)
}
