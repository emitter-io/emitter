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

package keygen

import (
	"fmt"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/errors"
	secmock "github.com/emitter-io/emitter/internal/provider/contract/mock"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/security/license"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExtendKey(t *testing.T) {
	license, _ := license.Parse(keygenTestLicense)
	tests := []struct {
		key           string
		channel       string
		access        uint8
		expires       time.Time
		contractValid bool
		contractFound bool
		err           *errors.Error
		expectAccess  uint8
	}{
		{
			key:           "DKs-8DXiPnaQjHm0ZwZPOBji-HsIExCF",
			channel:       "a/b/",
			contractValid: true,
			contractFound: true,
			access:        security.AllowAll,
			expectAccess:  security.AllowRead | security.AllowLoad,
		},
		{
			key:           "DKs-8DXiPnaQjHm0ZwZPOBji-HsIExCF",
			channel:       "a/b/",
			contractValid: true,
			contractFound: true,
			access:        security.AllowAll &^ security.AllowLoad,
			expectAccess:  security.AllowRead,
		},
		{
			key:           "Oad-avDCDdC-qPHLOANcUrDXm5eIEBFp",
			channel:       "a/b/",
			contractValid: true,
			contractFound: true,
			access:        security.AllowAll,
			err:           errors.ErrUnauthorized,
		},
	}
	for i, tc := range tests {
		name := fmt.Sprintf("case %v", i)
		t.Run(name, func(*testing.T) {
			provider := secmock.NewContractProvider()
			contract := new(secmock.Contract)
			contract.On("Validate", mock.Anything).Return(tc.contractValid)
			contract.On("Stats").Return(usage.NewMeter(0))
			provider.On("Get", mock.Anything).Return(contract, tc.contractFound)
			cipher, _ := license.Cipher()
			p := NewProvider(cipher, provider)

			channel, err := p.ExtendKey(tc.key, tc.channel, "ID", tc.access, tc.expires)
			if tc.err != nil {
				assert.Equal(t, tc.err, err, name)
				return
			}

			// Successful case
			assert.Nil(t, err, name)
			assert.NotNil(t, channel, name)
			assert.Equal(t, "a/b/ID/", string(channel.Channel))

			// Make sure the permissions are valid
			k, kerr := p.DecryptKey(string(channel.Key))
			assert.NoError(t, kerr)
			assert.Equal(t, tc.expectAccess, k.Permissions())
		})
	}
}

func TestCreateKey(t *testing.T) {
	license, _ := license.Parse("N7XxQbUEPxJ_RIj4muLUdLGYtR1kdKe2AAAAAAAAAAI")
	tests := []struct {
		key           string
		channel       string
		access        uint8
		expires       time.Time
		contractValid bool
		contractFound bool
		err           *errors.Error
	}{
		{
			key:           "xEbaDPaICEwVhgdnl2rg_1DWi_MAg_3B",
			channel:       "article1",
			contractValid: true,
			contractFound: true,
			err:           errors.ErrUnauthorized,
		},
		{
			key:           "xEbaDPaICEwVhgdnl2rg_1DWi_MAg_3B",
			channel:       "article1",
			contractValid: true,
			contractFound: true,
			err:           errors.ErrUnauthorized,
		},
		{
			key:           "8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR",
			channel:       "article1",
			contractValid: true,
			contractFound: false,
			err:           errors.ErrNotFound,
		},
		{
			key:           "8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR",
			channel:       "article1",
			contractValid: false,
			contractFound: true,
			err:           errors.ErrUnauthorized,
		},
		{
			key:           "8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR",
			channel:       "article1",
			contractValid: true,
			contractFound: true,
			err:           errors.ErrTargetInvalid,
		},
		{
			key:           "8GR6MtpL7Xut-pyogQMeS_gyxEA21BbR",
			channel:       "article1/",
			contractValid: true,
			contractFound: true,
		},
	}

	for i, tc := range tests {
		name := fmt.Sprintf("case %v", i)
		t.Run(name, func(*testing.T) {
			provider := secmock.NewContractProvider()
			contract := new(secmock.Contract)
			contract.On("Validate", mock.Anything).Return(tc.contractValid)
			contract.On("Stats").Return(usage.NewMeter(0))
			provider.On("Get", mock.Anything).Return(contract, tc.contractFound)
			cipher, _ := license.Cipher()
			p := NewProvider(cipher, provider)

			_, err := p.CreateKey(tc.key, tc.channel, tc.access, tc.expires)
			if tc.err != nil {
				assert.Equal(t, tc.err, err, name)
			} else {
				assert.Nil(t, err, name)
			}
		})
	}
}
