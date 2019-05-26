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

	"github.com/stretchr/testify/assert"
)

func TestNewV1(t *testing.T) {
	l := NewV1()
	assert.NotEqual(t, "", l.EncryptionKey)
	assert.Len(t, l.EncryptionKey, 22)

	c, err := l.Cipher()
	assert.NotNil(t, c)
	assert.NoError(t, err)

	text := l.String()
	assert.NotEqual(t, "", text)
	assert.Equal(t, ":1", text[len(text)-2:])

	out, err := parseV1(text[:len(text)-2])
	assert.NoError(t, err)
	assert.Equal(t, l, out)

	master, err := l.NewMasterKey(9)
	assert.NoError(t, err)
	assert.Equal(t, 9, int(master.Master()))
}

func TestParseV1(t *testing.T) {
	l, err := Parse("zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAFCDVAAAAAAI:1")
	assert.NoError(t, err)
	assert.Equal(t, uint32(0x3afc281d), l.Contract())
	assert.Equal(t, uint32(0x0), l.Signature())
	assert.Equal(t, uint32(0x1), l.Master())
}