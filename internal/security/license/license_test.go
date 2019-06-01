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

package license

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLicenseEncode(t *testing.T) {
	tests := []struct {
		license  License
		expected string
	}{
		{expected: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAFCDVAAAAAAI:1", license: &V1{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
			User:          989603869,
			Expires:       time.Unix(1600000000, 0),
			Type:          2,
		}},
		{expected: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI:1", license: &V1{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
			User:          989603869,
			Expires:       time.Unix(0, 0),
			Type:          2,
		}},
		{expected: "", license: &V1{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPT%",
		}},
		{expected: "CSAAAJ3Q8NcDAAA:2", license: &V2{
			User: 989603869,
		}},
	}

	for _, tc := range tests {
		output := tc.license.String()
		assert.EqualValues(t, tc.expected, output)

		if tc.expected != "" {
			_, err := Parse(tc.expected)
			assert.NoError(t, err)
		}
	}
}

func TestNewLicenseAndMaster(t *testing.T) {
	// license: RfBEIIFz1nNLf12JYRpoEUqFPLb3na0X_xbP_h3PM_CqDUVBGJfEV3WalW2maauQd48o-TcTM_61BfEsELfk0qMDqrCTswkB:2
	// secret:  wnLJv3TMhYTg6lLkGfQoazo1-k7gjFPk
	assert.NotPanics(t, func() {
		l, m := New()
		assert.NotEmpty(t, l)
		assert.NotEmpty(t, m)
	})
}

func TestParseLicense(t *testing.T) {
	tests := []struct {
		license  string
		expected License
		err      bool
	}{
		{license: "", err: true},
		{license: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI#", err: true},
		{license: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAFCDVAAAAAAI", expected: &V1{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
			User:          989603869,
			Expires:       time.Unix(1600000000, 0),
			Type:          2,
		}},
		{license: "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI9", expected: &V1{
			EncryptionKey: "zT83oDV0DWY5_JysbSTPTA",
			User:          989603869,
			Expires:       time.Unix(0, 0),
			Type:          2,
		}},
	}

	for _, tc := range tests {
		output, err := Parse(tc.license)
		assert.Equal(t, tc.err, err != nil)
		if !tc.err {
			assert.EqualValues(t, tc.expected, output)

			cipher, err := output.Cipher()
			assert.NoError(t, err)
			assert.NotNil(t, cipher)
		}
	}
}
