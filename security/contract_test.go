package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSingleContractProvider_Create(t *testing.T) {
	license, err := ParseLicense("zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI")
	if err != nil {
		t.Error(err)
	}

	contract := new(SingleContractProvider)
	contract.Create(license)

	assert.EqualValues(t, contract.Data.MasterID, 1)
	assert.EqualValues(t, contract.Data.Signature, license.Signature)
	assert.EqualValues(t, contract.Data.ID, license.Contract)
}

func TestSingleContractProvider_GetById(t *testing.T) {
	license, err := ParseLicense("zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI")
	if err != nil {
		t.Error(err)
	}

	contract := new(SingleContractProvider)
	contract.Create(license)
	contractByID := contract.GetByID(license.Contract)
	contractByWrongID := contract.GetByID(0)
	assert.NotNil(t, contractByID)
	assert.Nil(t, contractByWrongID)
}

/*
func TestSingleContractProvider_Validate(t *testing.T) {
	license, err := ParseLicense("zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI")
	if err != nil {
		t.Error(err)
	}

	contract := new(SingleContractProvider)
	contractData := contract.Create(license)

	assert.True(t, contractData.Validate(license.EncryptionKey))
}*/
