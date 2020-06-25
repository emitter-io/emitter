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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Benchmark_Time/new-8         	57209055	        20.7 ns/op	      16 B/op	       1 allocs/op
func Benchmark_Time(b *testing.B) {

	// Encode
	b.Run("new", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t := newTime(10, 20, nil)
			t.encode()
		}
	})
}

func TestNew(t *testing.T) {
	s1 := New(true, "")
	assert.IsType(t, new(Durable), s1)

	s2 := New(false, "")
	assert.IsType(t, new(Volatile), s2)
}

func TestTimeCodec(t *testing.T) {
	v1 := newTime(10, 50, []byte("hello"))
	enc := v1.encode()
	assert.Equal(t, int64(10), v1.AddTime())
	assert.Equal(t, int64(50), v1.DelTime())
	assert.Equal(t, []byte{
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, // 10
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, // 50
		0x68, 0x65, 0x6c, 0x6c, 0x6f, // hello
	}, []byte(enc))

	assert.Equal(t, 21, cap(v1))

	v2 := decodeValue(enc)
	assert.Equal(t, v1, v2)

}

func TestTime(t *testing.T) {
	v := newTime(Now(), Now(), []byte("hello"))
	assert.Equal(t, 21, len(v))
	assert.Equal(t, 21, cap(v))
	v.setValue([]byte("larger value"))
	assert.Equal(t, 28, len(v))
	v.setValue(nil)
	assert.Equal(t, 16, len(v))
}

// ------------------------------------------------------------------------------------

// Lock for the timer
var lock sync.Mutex

// RestoreClock restores the clock time
func restoreClock(clk clock) {
	lock.Lock()
	Now = clk
	lock.Unlock()
}

// SetClock sets the clock time for testing
func setClock(t int64) {
	lock.Lock()
	Now = func() int64 { return t }
	lock.Unlock()
}

// New creates a new CRDT set.
func mapOf(durable bool, entries ...func() (string, Value)) Map {
	m := make(map[string]Value, len(entries))
	for _, f := range entries {
		k, v := f()
		m[k] = v
	}

	if durable {
		return newDurableWith("", m)
	}
	return newVolatileWith(m)
}

type Action = func() (string, Value)

// T returns the entry constructor.
func T(k string, add, del int64, v ...string) Action {
	if len(v) == 0 {
		v = append(v, "")
	}

	value := newTime(add, del, []byte(v[0]))
	return func() (string, Value) {
		return k, value
	}
}

func equalSets(t *testing.T, expected, current Map) {
	expected.Range(nil, true, func(k string, v Value) bool {
		assert.Equal(t, v, current.Get(k))
		return true
	})
}

// newTime creates a new instance of time stamp.
func newTime(addTime, delTime int64, value []byte) Value {
	b := Value(make([]byte, 16, 16+len(value)))
	b.setAddTime(addTime)
	b.setDelTime(delTime)
	b.setValue(value)
	return b
}
