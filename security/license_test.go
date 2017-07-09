package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLicense(t *testing.T) {
	input := "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI"
	output, err := NewLicense(input)
	if err != nil {
		t.Error(err)
	}

	expect := License{
		EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
		Contract:      989603869,
		Expires:       time.Unix(0, 0),
		Type:          2,
	}

	assert.EqualValues(t, expect, *output)
}

func TestLicenseEncode(t *testing.T) {
	input := License{
		EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
		Contract:      989603869,
		Expires:       time.Unix(0, 0),
		Type:          2,
	}

	expect := "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI"
	output := input.String()

	assert.EqualValues(t, expect, output)
}
