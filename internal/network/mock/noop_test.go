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

package mock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoop(t *testing.T) {
	end := NewNoop()

	assert.Equal(t, "127.0.0.1", end.LocalAddr().String())
	assert.Equal(t, "127.0.0.1", end.RemoteAddr().String())
	assert.NoError(t, end.Close())
	assert.NoError(t, end.SetDeadline(time.Now()))
	assert.NoError(t, end.SetReadDeadline(time.Now()))
	assert.NoError(t, end.SetWriteDeadline(time.Now()))

	n, err := end.Read(nil)
	assert.Equal(t, 0, n)
	assert.NoError(t, err)

	n, err = end.Write([]byte{1, 2, 3})
	assert.Equal(t, 3, n)
	assert.NoError(t, err)
}
