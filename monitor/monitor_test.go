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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func BenchmarkMeasure(b *testing.B) {
	m := New()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Measure("abc", 15423)
	}
}

func BenchmarkEncode(b *testing.B) {
	m := New()
	for i := 0; i < 50; i++ {
		for j := 0; j < 100; j++ {
			m.Measure(fmt.Sprintf("%d", j), int32(i))
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Snapshot()
	}
}

func BenchmarkRuntime(b *testing.B) {
	m := New()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.MeasureRuntime()
	}
}

func TestMeasureElapsed(t *testing.T) {
	m := New()

	measureDelay(m)
	elapsed := m.Get("a").Max()
	assert.NotZero(t, elapsed)
}

func measureDelay(m *Monitor) {
	defer m.MeasureElapsed("a", time.Now())
	time.Sleep(1 * time.Millisecond)
}

func TestMonitorTag(t *testing.T) {
	m := New()
	m.Tag("a", "roman")

	assert.Equal(t, "roman", m.Get("a").Tag())
}

func TestMeasureRuntime(t *testing.T) {
	m := New()
	m.MeasureRuntime()

	assert.NotZero(t, m.Get("go.procs").Max())
}

func TestHistogramEncodeMany(t *testing.T) {
	m := New()

	for i := 0; i < 1000; i++ {
		for j := 0; j < 100; j++ {
			m.Measure(fmt.Sprintf("%d", j), rand.Int31n(10000))
		}
	}

	v := m.Snapshot()
	assert.True(t, len(v) > 1000)
}

func TestHistogram(t *testing.T) {
	m := New()

	for i := 0; i < 5000; i++ {
		m.MeasureElapsed("b", time.Unix(0, 0))
		m.Measure("a", int32(i))
	}

	// Snapshot
	v := m.Snapshot()
	assert.True(t, len(v) > 50)

	// Restore
	h, err := Restore(v)
	assert.NoError(t, err)
	assert.Len(t, h, 2)
	assert.Equal(t, 5000, h[0].Count())
	assert.Equal(t, 5000, h[1].Count())
}
