// +build !js

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

package stats

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Measurer represents a monitoring contract.
type Measurer interface {
	Snapshotter
	Measure(name string, value int32)
	MeasureElapsed(name string, start time.Time)
	MeasureRuntime()
	Tag(name, tag string)
}

// Metric maintains a combination of a gauge and a statistically-significant selection
// of the values from a stream. This is essentially a combination of a histogram, gauge
// and a counter.
type Metric struct {
	sync.Mutex
	data   sample // The sample used to build a histogram
	count  int32  // The number of samples observed
	create int64  // The first updated time
	name   string // The name of the metric
	tag    string // The tag of the metric (e.g.: IP Address)
}

const (
	reservoirSize = 1024
)

// NewMetric creates a new metric.
func NewMetric(name string) *Metric {
	return &Metric{
		name:   name,
		data:   make([]int32, reservoirSize, reservoirSize),
		create: time.Now().Unix(),
	}
}

// Reset clears all samples and resets the metric.
func (m *Metric) Reset() {
	m.Lock()
	defer m.Unlock()

	atomic.StoreInt32(&m.count, 0)
	m.create = time.Now().Unix()
}

// Name returns the name of the histogram.
func (m *Metric) Name() string {
	return m.name
}

// Tag returns the associated tag of the metric.
func (m *Metric) Tag() string {
	m.Lock()
	defer m.Unlock()
	return m.tag
}

// Window returns start and end time of the histogram.
func (m *Metric) Window() (time.Time, time.Time) {
	return time.Unix(m.create, 0), time.Now()
}

// sample returns the usable sample
func (m *Metric) sample() sample {
	count := m.count
	if count > reservoirSize {
		count = reservoirSize
	}
	return m.data[:count]
}

// Count returns the number of samples recorded, which may exceed the
// reservoir size.
func (m *Metric) Count() int {
	m.Lock()
	defer m.Unlock()
	return int(m.count)
}

// Max returns the maximum value in the sample, which may not be the maximum
// value ever to be part of the sample.
func (m *Metric) Max() int {
	m.Lock()
	defer m.Unlock()
	return m.sample().Max()
}

// Mean returns the mean of the values in the sample.
func (m *Metric) Mean() float64 {
	m.Lock()
	defer m.Unlock()
	return m.sample().Mean()
}

// Min returns the minimum value in the sample, which may not be the minimum
// value ever to be part of the sample.
func (m *Metric) Min() int {
	m.Lock()
	defer m.Unlock()
	return m.sample().Min()
}

// Quantile returns a slice of arbitrary quantiles of the sample.
func (m *Metric) Quantile(quantiles ...float64) []float64 {
	m.Lock()
	defer m.Unlock()
	return m.sample().Quantile(quantiles...)
}

// Snapshot returns a read-only copy of the sample.
func (m *Metric) Snapshot() *Snapshot {
	m.Lock()
	defer m.Unlock()

	// Snapshot the data
	sample := m.sample()
	dest := make([]int32, len(sample))
	copy(dest, sample)
	return &Snapshot{
		Metric: m.name,
		Label:  m.tag,
		T0:     m.create,
		T1:     time.Now().Unix(),
		Amount: m.count,
		Sample: dest,
	}
}

// StdDev returns the standard deviation of the values in the sample.
func (m *Metric) StdDev() float64 {
	m.Lock()
	defer m.Unlock()
	return m.sample().StdDev()
}

// Variance returns the variance of the values in the sample.
func (m *Metric) Variance() float64 {
	m.Lock()
	defer m.Unlock()
	return m.sample().Variance()
}

// Rate returns a operation per second rate since the creation of the metric.
func (m *Metric) Rate() float64 {
	t0, t1 := m.Window()
	return float64(m.Count()) / float64(t1.Sub(t0).Seconds())
}

// Update samples a new value into the metric.
func (m *Metric) Update(v int32) {
	count := atomic.AddInt32(&m.count, 1)
	if count <= reservoirSize {
		m.Lock()
		m.data[count-1] = v
		m.Unlock()
		return
	}

	if r := int(rand.Int31n(count)); r < reservoirSize {
		m.Lock()
		m.data[r] = v
		m.Unlock()
		return
	}
}

// UpdateTag updates the associated metric tag.
func (m *Metric) UpdateTag(tag string) {
	m.Lock()
	m.tag = tag
	m.Unlock()
}
