/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
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

package message

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID_NewID(t *testing.T) {
	next = math.MaxUint32
	id := NewID(Ssid{1, share, 3})

	assert.Zero(t, next)
	assert.True(t, id.Time() > 1527819700)

	id.SetTime(offset + 1)
	assert.Equal(t, int64(offset+1), id.Time())
}

func TestID_NewPrefix(t *testing.T) {
	id1 := NewID(Ssid{1, 2, 3})
	id2 := NewPrefix(Ssid{1, 2, 3}, 123)

	assert.Equal(t, id1[:4], id2[:4])
}

func TestID_HasPrefix(t *testing.T) {
	next = 0
	id := NewID(Ssid{1, 2, 3})

	assert.True(t, id.HasPrefix(Ssid{1, 2}, 0))
	assert.False(t, id.HasPrefix(Ssid{1, 2}, 2527784701635600500)) // After Sunday, February 6, 2050 6:25:01.636 PM this test will fail :-)
	assert.False(t, id.HasPrefix(Ssid{1, 3}, 0))
}

func TestID_Match(t *testing.T) {
	id := NewID(Ssid{1, 2, 3, 4})

	assert.NotEmpty(t, id)
	assert.True(t, id.Match(Ssid{1, 2, 3, 4}, 0, math.MaxInt64))
	assert.True(t, id.Match(Ssid{1, 2, 3}, 0, math.MaxInt64))
	assert.True(t, id.Match(Ssid{1, 2}, 0, math.MaxInt64))
	assert.False(t, id.Match(Ssid{1, 2, 3, 5}, 0, math.MaxInt64))
	assert.False(t, id.Match(Ssid{1, 5}, 0, math.MaxInt64))
	assert.False(t, id.Match(Ssid{1, 5}, 0, math.MaxInt64))
	assert.True(t, id.Match(Ssid{1, 2}, 0, 2527784701635600500))
	assert.False(t, id.Match(Ssid{1, 2}, 2527784701635600500, math.MaxInt64))
	assert.False(t, id.Match(Ssid{2, 2, 3, 4}, 0, math.MaxInt64))
	assert.False(t, id.Match(Ssid{2, 2, 3, 4, 5}, 0, math.MaxInt64))
}

func TestID_Ssid(t *testing.T) {
	in := Ssid{1, 2, 3, 4, 5, 6}
	id := NewID(in)

	assert.Equal(t, in[0], id.Contract())
	assert.Equal(t, in, id.Ssid())
}

func BenchmarkID_New(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	ssid := Ssid{1, 2}
	for n := 0; n < b.N; n++ {
		NewID(ssid)
	}
}

func BenchmarkID_Match(b *testing.B) {
	id := NewID(Ssid{1, 2, 3, 4})
	ssid := Ssid{1, 2, 3, 4}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		id.Match(ssid, 0, math.MaxInt64)
	}
}
