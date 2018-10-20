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

package stats

import (
	"github.com/golang/snappy"
	"github.com/kelindar/binary"
)

// Snapshotter represents a snapshotting contract.
type Snapshotter interface {
	Snapshot() []byte
}

// Restore restores a snapshot into a read-only histogram format.
func Restore(encoded []byte) (snapshots Snapshots, err error) {
	var decoded []byte
	if decoded, err = snappy.Decode(decoded, encoded); err == nil {
		err = binary.Unmarshal(decoded, &snapshots)
	}
	return
}

// Snapshot is a read-only copy of another Sample.
type Snapshot struct {
	Metric string
	Label  string
	Sample sample
	Amount int32
	T0     int64
	T1     int64
}

// Name returns the name of the metric.
func (s *Snapshot) Name() string {
	return s.Metric
}

// Tag returns the associated tag of the metric.
func (s *Snapshot) Tag() string {
	return s.Label
}

// Window returns start and end time of the metric.
func (s *Snapshot) Window() (int64, int64) {
	return s.T0, s.T1
}

// Count returns the count of inputs at the time the snapshot was taken.
func (s *Snapshot) Count() int {
	return int(s.Amount)
}

// Max returns the maximal value at the time the snapshot was taken.
func (s *Snapshot) Max() int {
	return s.Sample.Max()
}

// Mean returns the mean value at the time the snapshot was taken.
func (s *Snapshot) Mean() float64 {
	return s.Sample.Mean()
}

// Min returns the minimal value at the time the snapshot was taken.
func (s *Snapshot) Min() int {
	return s.Sample.Min()
}

// Sum returns the arithmetic sum of all of the values in the snapshot.
func (s *Snapshot) Sum() int {
	return int(s.Sample.Sum())
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

// Rate returns a operation per second rate over the time window.
func (s *Snapshot) Rate() float64 {
	return float64(s.Amount) / float64(s.T1-s.T0)
}

// Merge merges two snapshots together.
func (s *Snapshot) Merge(other Snapshot) {
	s.Sample = append(s.Sample, other.Sample...)
	s.Amount = s.Amount + other.Amount
	if other.T0 < s.T0 {
		s.T0 = other.T0
	}
	if other.T1 > s.T1 {
		s.T1 = other.T1
	}
}

// Snapshots represents a set of snapshots.
type Snapshots []Snapshot

// ToMap converts the set of snapshots to a map.
func (snapshots Snapshots) ToMap() map[string]Snapshot {
	out := make(map[string]Snapshot)
	for _, s := range snapshots {
		out[s.Name()] = s
	}
	return out
}

// Merge merges two sets of snapshots together.
func (snapshots *Snapshots) Merge(others Snapshots) {
	m0 := snapshots.ToMap()
	m1 := others.ToMap()

	// Merge existing snapshots
	for _, s := range *snapshots {
		if other, ok := m1[s.Name()]; ok {
			s.Merge(other)
		}
	}

	// Append missing snapshots
	for _, s := range others {
		if _, ok := m0[s.Name()]; !ok {
			*snapshots = append(*snapshots, s)
		}
	}
}
