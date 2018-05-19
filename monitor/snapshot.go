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
	"time"
)

// Snapshot is a read-only copy of another Sample.
type Snapshot struct {
	Metric string
	Sample sample
	VCount int32
	VMin   int64
	VMax   int64
	Create int64
	Update int64
}

// newSnapshot creates a new snapshot from a metric.
func newSnapshot(metric *Metric) *Snapshot {
	sample := make([]int64, len(metric.sample))
	copy(sample, metric.sample)
	return &Snapshot{
		Metric: metric.name,
		Create: metric.create,
		Update: metric.update,
		VCount: metric.count,
		VMin:   metric.min,
		VMax:   metric.max,
		Sample: sample,
	}
}

// Name returns the name of the metric.
func (s *Snapshot) Name() string {
	return s.Metric
}

// Window returns start and end time of the metric.
func (s *Snapshot) Window() (time.Time, time.Time) {
	return time.Unix(s.Create, 0), time.Unix(s.Update, 0)
}

// Count returns the count of inputs at the time the snapshot was taken.
func (s *Snapshot) Count() int {
	return int(s.VCount)
}

// Max returns the maximal value at the time the snapshot was taken.
func (s *Snapshot) Max() int64 {
	return s.VMax
}

// Mean returns the mean value at the time the snapshot was taken.
func (s *Snapshot) Mean() float64 {
	return s.Sample.Mean()
}

// Min returns the minimal value at the time the snapshot was taken.
func (s *Snapshot) Min() int64 {
	return s.VMin
}

// Quantile returns a slice of arbitrary quantiles of the sample.
func (s *Snapshot) Quantile(quantiles ...float64) []float64 {
	return s.Sample.Quantile(quantiles...)
}

// StdDev returns the standard deviation of values at the time the snapshot was
// taken.
func (s *Snapshot) StdDev() float64 {
	return s.Sample.StdDev()
}

// Variance returns the variance of values at the time the snapshot was taken.
func (s *Snapshot) Variance() float64 {
	return s.Sample.Variance()
}
