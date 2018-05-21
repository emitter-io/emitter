// +build !js

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

func BenchmarkMetricSnapshot(b *testing.B) {
	h := NewMetric("x")
	for i := int32(0); i < 50000; i++ {
		h.Update(i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Snapshot()
	}
}

func BenchmarkMetricUpdate(b *testing.B) {
	m := NewMetric("x")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Update(15423)
	}
}

func TestMetric(t *testing.T) {
	h := NewMetric("x")
	for i := int32(0); i < 100; i++ {
		h.Update(i)
	}

	h.UpdateTag("test")
	assert.Equal(t, "test", h.Tag())

	// Create a snapshot
	assert.Equal(t, 100, h.Count())
	assert.Equal(t, 99, h.Max())
	assert.Equal(t, 0, h.Min())
	assert.True(t, h.Mean() > 49)
	assert.True(t, h.StdDev() > 28)
	assert.Equal(t, "x", h.Name())
	assert.Equal(t, float64(49.5), h.Quantile(50)[0])
	assert.Equal(t, 833.25, h.Variance())

	t0, t1 := h.Window()
	assert.NotEqual(t, time.Unix(0, 0), t0)
	assert.NotEqual(t, time.Unix(0, 0), t1)

	h.Reset()
	assert.Equal(t, 0, h.Count())
	assert.Equal(t, 0, h.Max())
}
