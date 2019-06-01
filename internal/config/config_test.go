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

package config

import (
	"os"
	"strings"
	"testing"

	"github.com/emitter-io/config/dynamo"
	"github.com/stretchr/testify/assert"
)

func Test_NewDefaut(t *testing.T) {
	c := NewDefault().(*Config)
	assert.Equal(t, ":8080", c.ListenAddr)
	//assert.Nil(t, c.Vault())

	tls, _, ok := c.Certificate()
	assert.Nil(t, tls)
	assert.False(t, ok)
}

func Test_Addr(t *testing.T) {
	c := &Config{
		ListenAddr: "private",
	}

	addr := c.Addr()
	assert.True(t, strings.HasSuffix(addr.String(), ":8080"))
}

func Test_AddrInvalid(t *testing.T) {
	assert.Panics(t, func() {
		c := &Config{ListenAddr: "g3ew235wgs"}
		c.Addr()
	})
}

func Test_New(t *testing.T) {
	c := New("test.conf", dynamo.NewProvider())
	defer os.Remove("test.conf")

	assert.NotNil(t, c)
}
