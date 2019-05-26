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

func TestNewV2(t *testing.T) {
	l := NewV2()
	assert.NotEqual(t, "", l.EncryptionKey)
	assert.Len(t, l.EncryptionKey, 32)

	c, err := l.Cipher()
	assert.NotNil(t, c)
	assert.NoError(t, err)

	text := l.String()
	assert.NotEqual(t, "", text)
	assert.Equal(t, ":2", text[len(text)-2:])

	out, err := parseV2(text[:len(text)-2])
	assert.NoError(t, err)
	assert.Equal(t, l, out)

	master, err := l.NewMasterKey(9)
	assert.NoError(t, err)
	assert.Equal(t, 9, int(master.Master()))
}

func TestParseV2(t *testing.T) {
	l, err := Parse("RfBEIO9PA3cczC6bZQGnEeX8zbgKhm5Gw4ZlJSFJsaChuGStGCKNZ-8LTxKwiD7wK8EOAhrmleUY7PbLHrmCkokB0NaYgAEB:2")
	assert.NoError(t, err)
	assert.Equal(t, uint32(0x11248139), l.Contract())
	assert.Equal(t, uint32(0x10062b50), l.Signature())
	assert.Equal(t, uint32(0x1), l.Master())
}

func TestParseV2_Invalid(t *testing.T) {
	_, err := Parse("``````````:2")
	assert.Error(t, err)

	_, err = Parse("xxxxxx:2")
	assert.Error(t, err)
}
