package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseLicense(t *testing.T) {
	tests := []struct {
		license  string
		expected License
		err      bool
	}{
		{license: "", err: true},
		{license: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI#", err: true},
		{license: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAFCDVAAAAAAI", expected: License{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
			Contract:      989603869,
			Expires:       time.Unix(1600000000, 0),
			Type:          2,
		}},
		{license: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI9", expected: License{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
			Contract:      989603869,
			Expires:       time.Unix(0, 0),
			Type:          2,
		}},
	}

	for _, tc := range tests {
		output, err := ParseLicense(tc.license)
		assert.Equal(t, tc.err, err != nil)
		if !tc.err {
			assert.EqualValues(t, tc.expected, *output)

			cipher, err := output.Cipher()
			assert.NoError(t, err)
			assert.NotNil(t, cipher)
		}
	}
}

func TestLicenseEncode(t *testing.T) {
	tests := []struct {
		license  License
		expected string
	}{
		{expected: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAFCDVAAAAAAI", license: License{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
			Contract:      989603869,
			Expires:       time.Unix(1600000000, 0),
			Type:          2,
		}},
		{expected: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI", license: License{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
			Contract:      989603869,
			Expires:       time.Unix(0, 0),
			Type:          2,
		}},
		{expected: "", license: License{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPT%",
		}},
	}

	for _, tc := range tests {
		output := tc.license.String()
		assert.EqualValues(t, tc.expected, output)
	}
}
