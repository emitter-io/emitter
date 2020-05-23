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
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/golang/snappy"
	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	defer restoreClock(Now)

	setClock(0)
	state := NewWith(map[string]Time{"A": {AddTime: 10, DelTime: 50}})

	// Encode
	enc, err := binary.Marshal(state)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x1, 0x14, 0x64, 0x1, 0x41}, enc)

	// Decode
	dec := New()
	err = binary.Unmarshal(enc, dec)
	assert.NoError(t, err)
	assert.Equal(t, state, dec)
}

// 15852470 -> 3632341 bytes, 22.91%
func TestSizeMarshal(t *testing.T) {
	state, size := loadTestData(t)

	// Encode
	enc, err := binary.Marshal(state)
	assert.NoError(t, err)

	fmt.Printf("%d -> %d bytes, %.2f%% \n", size, len(enc), float64(len(enc))/float64(size)*100)
	assert.Greater(t, 20000000, len(enc))

	// Decode
	out := New()
	err = binary.Unmarshal(enc, out)
	assert.NoError(t, err)
	assert.Equal(t, 100000, len(out.data))
}

// Benchmark_Marshal/encode-8         	      85	  12094106 ns/op	11090183 B/op	      19 allocs/op
// Benchmark_Marshal/decode-8         	      74	  15932831 ns/op	10636099 B/op	    3936 allocs/op
func Benchmark_Marshal(b *testing.B) {
	state, _ := loadTestData(b)

	// Encode
	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			binary.Marshal(state)
		}
	})

	// Decode
	enc, err := binary.Marshal(state)
	assert.NoError(b, err)
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			state := New()
			binary.Unmarshal(enc, state)
		}
	})
}

func loadTestData(t assert.TestingT) (state *Set, size int) {
	state = New()
	buf, err := ioutil.ReadFile("test.bin")
	assert.NoError(t, err)

	decoded, err := snappy.Decode(nil, buf)
	assert.NoError(t, err)

	err = binary.Unmarshal(decoded, &state.data)
	assert.NoError(t, err)
	size = len(decoded)
	return
}
