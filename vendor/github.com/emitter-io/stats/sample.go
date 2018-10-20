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
	"math"
	"reflect"
	"sort"

	"github.com/kelindar/binary"
	"github.com/kelindar/binary/sorted"
)

// Sample represents a sample window
type sample sorted.Int32s

func (s sample) Len() int           { return len(s) }
func (s sample) Less(i, j int) bool { return s[i] < s[j] }
func (s sample) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// GetBinaryCodec retrieves a custom binary codec.
func (s *sample) GetBinaryCodec() binary.Codec {
	return sorted.IntsCodecAs(reflect.TypeOf(sample{}), 4)
}

// StdDev returns the standard deviation of the sample.
func (s sample) StdDev() float64 {
	return math.Sqrt(s.Variance())
}

// Sum returns the sum of the sample.
func (s sample) Sum() (sum int64) {
	for _, v := range s {
		sum += int64(v)
	}
	return
}

// Variance returns the variance of the sample.
func (s sample) Variance() float64 {
	if 0 == len(s) {
		return 0.0
	}

	m := s.Mean()
	var sum float64
	for _, v := range s {
		d := float64(v) - m
		sum += d * d
	}
	return sum / float64(len(s))
}

// Variance returns the mean of the sample.
func (s sample) Mean() float64 {
	if 0 == len(s) {
		return 0.0
	}

	return float64(s.Sum()) / float64(len(s))
}

// Min returns the minimum value of the sample.
func (s sample) Min() int {
	if 0 == len(s) {
		return 0
	}

	var min int32 = math.MaxInt32
	for _, v := range s {
		if min > v {
			min = v
		}
	}
	return int(min)
}

// Max returns the maximum value of the sample.
func (s sample) Max() int {
	if 0 == len(s) {
		return 0
	}

	var max int32 = math.MinInt32
	for _, v := range s {
		if max < v {
			max = v
		}
	}
	return int(max)
}

// Quantiles returns a slice of arbitrary quantiles of the sample.
func (s sample) Quantile(quantiles ...float64) []float64 {
	scores := make([]float64, len(quantiles))
	size := len(s)
	if size > 0 {
		sort.Sort(s)
		for i, quantile := range quantiles {
			pos := (quantile / 100) * float64(size+1)
			if pos < 1.0 {
				scores[i] = float64(s[0])
			} else if pos >= float64(size) {
				scores[i] = float64(s[size-1])
			} else {
				lower := float64(s[int(pos)-1])
				upper := float64(s[int(pos)])
				scores[i] = lower + (pos-math.Floor(pos))*(upper-lower)
			}
		}
	}
	return scores
}
