/**********************************************************************************
* Copyright (c) 2009-2020 Misakai Ltd.
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

package crdt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Benchmark_Time(b *testing.B) {
	t := Time{10, 20}

	// Encode
	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t.Encode()
		}
	})

	// Decode
	enc := t.Encode()
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			decodeTime(enc)
		}
	})
}

func TestNew(t *testing.T) {
	s1 := New(true)
	assert.IsType(t, new(Durable), s1)

	s2 := New(false)
	assert.IsType(t, new(Volatile), s2)
}

func TestTimeCodec(t *testing.T) {
	v1 := Time{AddTime: 10, DelTime: 50}
	enc := v1.Encode()
	assert.Equal(t, []byte{0x14, 0x64}, []byte(enc))

	v2 := decodeTime(enc)
	assert.Equal(t, v1, v2)
}
