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

	"github.com/stretchr/testify/assert"
)

func TestSampleZero(t *testing.T) {
	var s sample
	assert.Zero(t, s.Min())
	assert.Zero(t, s.Mean())
	assert.Zero(t, s.StdDev())
	assert.Zero(t, s.Variance())
	assert.Zero(t, s.Quantile(50)[0])
}

func TestQuantileZero(t *testing.T) {
	var s sample
	for i := int64(0); i < 10000; i++ {
		s = append(s, 0)
	}

	assert.Zero(t, s.Quantile(0.0001)[0])
	assert.Zero(t, s.Quantile(50)[0])
	assert.Zero(t, s.Quantile(5000)[0])
}

func TestQuantiles(t *testing.T) {
	var s sample
	for i := int64(0); i < 10000; i++ {
		s = append(s, i/100)
	}

	assert.Equal(t, float64(49.5), s.Quantile(50)[0])
}
