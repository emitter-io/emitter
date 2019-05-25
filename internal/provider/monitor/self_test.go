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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type snapshot string

func (s snapshot) Snapshot() []byte {
	return []byte(s)
}

func TestSelf(t *testing.T) {
	r := snapshot("test")
	cfg := map[string]interface{}{
		"interval": float64(100),
		"channel":  "chan",
	}

	s := NewSelf(r, func(c string, v []byte) {
		assert.True(t, strings.HasPrefix(c, "chan/"))
		assert.Equal(t, "test", string(v))
	})

	err := s.Configure(cfg)
	assert.NoError(t, err)
	assert.Equal(t, "self", s.Name())

	errClose := s.Close()
	assert.NoError(t, errClose)
}
