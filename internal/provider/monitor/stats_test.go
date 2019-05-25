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

package monitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoop_Configure(t *testing.T) {
	s := NewNoop()
	err := s.Configure(nil)
	assert.NoError(t, err)
}

func TestNoop_Name(t *testing.T) {
	s := new(Noop)
	assert.Equal(t, "noop", s.Name())
}

func TestNoop_Close(t *testing.T) {
	s := new(Noop)
	err := s.Close()
	assert.NoError(t, err)
}
