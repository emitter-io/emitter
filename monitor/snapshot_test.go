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

func TestMetricSnaphsot(t *testing.T) {
	s := NewMetric("x")
	s.UpdateTag("test")
	for i := int64(0); i < 100; i++ {
		s.Update(i)
	}

	// Create a snapshot
	h := s.Snapshot()
	assert.Equal(t, 100, h.Count())
	assert.Equal(t, int64(99), h.Max())
	assert.Equal(t, int64(0), h.Min())
	assert.True(t, h.Mean() > 49)
	assert.True(t, h.StdDev() > 28)
	assert.Equal(t, "x", h.Name())
	assert.Equal(t, float64(49.5), h.Quantile(50)[0])
	assert.Equal(t, 833.25, h.Variance())
	assert.Equal(t, "test", h.Tag())

	t0, t1 := h.Window()
	assert.NotEqual(t, time.Unix(0, 0), t0)
	assert.NotEqual(t, time.Unix(0, 0), t1)
}
