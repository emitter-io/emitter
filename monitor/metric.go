/**********************************************************************************
* Copyright (c) 2009-2018 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more detailm.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/

package monitor

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// Metric maintains a combination of a gauge and a statistically-significant selection
// of the values from a stream. This is essentially a combination of a histogram, gauge
// and a counter.
type Metric struct {
	sync.RWMutex
	sample sample // The sample used to build a histogram
	count  int32  // The number of samples observed
	min    int64  // The minimum value observed
	max    int64  // The maximum value observed
	update int64  // The last updated time
	create int64  // The first updated time
	name   string // The name of the metric
}

const (
	reservoirSize = 1024
)

// NewMetric creates a new metric.
func NewMetric(name string) *Metric {
	return &Metric{
		name:   name,
		sample: make([]int64, 0, reservoirSize),
		create: time.Now().Unix(),
		min:    math.MaxInt64,
	}
}

// Reset clears all samples and resets the metric.
func (m *Metric) Reset() {
	m.Lock()
	defer m.Unlock()

	m.count = 0
	m.min = math.MaxInt64
	m.max = 0
	m.sample = m.sample[:0]
}

// Name returns the name of the histogram.
func (m *Metric) Name() string {
	return m.name
}

// Window returns start and end time of the histogram.
func (m *Metric) Window() (time.Time, time.Time) {
	return time.Unix(m.create, 0), time.Unix(m.update, 0)
}

// Count returns the number of samples recorded, which may exceed the
// reservoir size.
func (m *Metric) Count() int {
	m.RLock()
	defer m.RUnlock()

	return int(m.count)
}

// Max returns the maximum value in the sample, which may not be the maximum
// value ever to be part of the sample.
func (m *Metric) Max() int64 {
	m.RLock()
	defer m.RUnlock()

	return m.sample.Max()
}

// Mean returns the mean of the values in the sample.
func (m *Metric) Mean() float64 {
	m.RLock()
	defer m.RUnlock()

	return m.sample.Mean()
}

// Min returns the minimum value in the sample, which may not be the minimum
// value ever to be part of the sample.
func (m *Metric) Min() int64 {
	m.RLock()
	defer m.RUnlock()

	return m.sample.Min()
}

// Quantile returns a slice of arbitrary quantiles of the sample.
func (m *Metric) Quantile(quantiles ...float64) []float64 {
	m.RLock()
	defer m.RUnlock()

	return m.sample.Quantile(quantiles...)
}

// Snapshot returns a read-only copy of the sample.
func (m *Metric) Snapshot() *Snapshot {
	m.RLock()
	defer m.RUnlock()

	return newSnapshot(m)
}

// StdDev returns the standard deviation of the values in the sample.
func (m *Metric) StdDev() float64 {
	m.RLock()
	defer m.RUnlock()

	return m.sample.StdDev()
}

// Variance returns the variance of the values in the sample.
func (m *Metric) Variance() float64 {
	m.Lock()
	defer m.Unlock()

	return m.sample.Variance()
}

// Update samples a new value into the metric.
func (m *Metric) Update(v int64) {
	now := time.Now().Unix()
	m.Lock()
	m.count++
	m.update = now
	if len(m.sample) < reservoirSize {
		m.sample = append(m.sample, v)
	} else if r := rand.Int31n(m.count); r < int32(len(m.sample)) {
		m.sample[int(r)] = v
	}

	if m.min > v {
		m.min = v
	}

	if m.max < v {
		m.max = v
	}
	m.Unlock()
}
