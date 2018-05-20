/**********************************************************************************
* Copyright (c) 2009-2018 Misakai Ltd.
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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoop(t *testing.T) {
	m := NewNoop()
	m.Measure("a", 1)
	m.MeasureElapsed("b", time.Now())
	m.MeasureRuntime()
	m.Tag("a", "b")
	assert.NotNil(t, m)

	b := m.Snapshot()
	assert.Empty(t, b)
}
