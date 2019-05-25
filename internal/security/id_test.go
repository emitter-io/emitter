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

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	defer func(n uint64) { next = n }(next)

	next = 0
	i1 := NewID()
	i2 := NewID()

	assert.Equal(t, ID(1), i1)
	assert.Equal(t, ID(2), i2)
}

func TestIDToString(t *testing.T) {
	defer func(n uint64) { next = n }(next)

	next = 0
	i1 := NewID()
	i2 := NewID()

	assert.Equal(t, "01", i1.String())
	assert.Equal(t, "02", i2.String())
}

func TestIDToUnique(t *testing.T) {
	defer func(n uint64) { next = n }(next)

	next = 0
	i1 := NewID()
	i2 := NewID()

	assert.Equal(t, "F45JPXDSXVRWBUKTDNCCM4PGQI", i1.Unique(123, "hello"))
	assert.Equal(t, "XCFU2OA7OO2COPZOJ5VA6GS6BM", i2.Unique(123, "hello"))
}
