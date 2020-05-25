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
	"sync"
	"testing"

	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
)

func TestVolatileAddContains(t *testing.T) {
	testStr := "ABCD"

	lww := NewVolatile()
	assert.False(t, lww.Contains(testStr))

	lww.Add(testStr)
	assert.True(t, lww.Contains(testStr))

	entry := lww.Get(testStr)
	assert.True(t, entry.IsAdded())
	assert.False(t, entry.IsRemoved())
	assert.False(t, entry.IsZero())
}

func TestVolatileAddRemoveContains(t *testing.T) {
	defer restoreClock(Now)

	setClock(0)

	lww := NewVolatile()
	testStr := "object2"

	lww.Add(testStr)

	setClock(100)
	lww.Remove(testStr)

	assert.False(t, lww.Contains(testStr))

	entry := lww.data[testStr]
	assert.False(t, entry.IsAdded())
	assert.True(t, entry.IsRemoved())
	assert.False(t, entry.IsZero())
}

func TestVolatileMerge(t *testing.T) {
	var T = func(add, del int64) Time {
		return Time{AddTime: add, DelTime: del}
	}

	for _, tc := range []struct {
		lww1, lww2, expected, delta *Volatile
		valid, invalid              []string
	}{
		{
			lww1:     newVolatileWith(map[string]Time{"A": T(10, 0), "B": T(20, 0)}),
			lww2:     newVolatileWith(map[string]Time{"A": T(0, 20), "B": T(0, 20)}),
			expected: newVolatileWith(map[string]Time{"A": T(10, 20), "B": T(20, 20)}),
			delta:    newVolatileWith(map[string]Time{"A": T(0, 20), "B": T(0, 20)}),
			valid:    []string{"B"},
			invalid:  []string{"A"},
		},
		{
			lww1:     newVolatileWith(map[string]Time{"A": T(10, 0), "B": T(20, 0)}),
			lww2:     newVolatileWith(map[string]Time{"A": T(0, 20), "B": T(10, 0)}),
			expected: newVolatileWith(map[string]Time{"A": T(10, 20), "B": T(20, 0)}),
			delta:    newVolatileWith(map[string]Time{"A": T(0, 20)}),
			valid:    []string{"B"},
			invalid:  []string{"A"},
		},
		{
			lww1:     newVolatileWith(map[string]Time{"A": T(30, 0), "B": T(20, 0)}),
			lww2:     newVolatileWith(map[string]Time{"A": T(20, 0), "B": T(10, 0)}),
			expected: newVolatileWith(map[string]Time{"A": T(30, 0), "B": T(20, 0)}),
			delta:    NewVolatile(),
			valid:    []string{"A", "B"},
			invalid:  []string{},
		},
		{
			lww1:     newVolatileWith(map[string]Time{"A": T(10, 0), "B": T(0, 20)}),
			lww2:     newVolatileWith(map[string]Time{"C": T(10, 0), "D": T(0, 20)}),
			expected: newVolatileWith(map[string]Time{"A": T(10, 0), "B": T(0, 20), "C": T(10, 0), "D": T(0, 20)}),
			delta:    newVolatileWith(map[string]Time{"C": T(10, 0), "D": T(0, 20)}),
			valid:    []string{"A", "C"},
			invalid:  []string{"B", "D"},
		},
		{
			lww1:     newVolatileWith(map[string]Time{"A": T(10, 0), "B": T(30, 0)}),
			lww2:     newVolatileWith(map[string]Time{"A": T(20, 0), "B": T(20, 0)}),
			expected: newVolatileWith(map[string]Time{"A": T(20, 0), "B": T(30, 0)}),
			delta:    newVolatileWith(map[string]Time{"A": T(20, 0)}),
			valid:    []string{"A", "B"},
			invalid:  []string{},
		},
		{
			lww1:     newVolatileWith(map[string]Time{"A": T(0, 10), "B": T(0, 30)}),
			lww2:     newVolatileWith(map[string]Time{"A": T(0, 20), "B": T(0, 20)}),
			expected: newVolatileWith(map[string]Time{"A": T(0, 20), "B": T(0, 30)}),
			delta:    newVolatileWith(map[string]Time{"A": T(0, 20)}),
			valid:    []string{},
			invalid:  []string{"A", "B"},
		},
	} {

		tc.lww1.Merge(tc.lww2)
		assert.Equal(t, tc.expected, tc.lww1, "Merged set is not the same")
		assert.Equal(t, tc.delta, tc.lww2, "Delta set is not the same")

		for _, obj := range tc.valid {
			assert.True(t, tc.lww1.Contains(obj), fmt.Sprintf("expected merged set to contain %v", obj))
		}

		for _, obj := range tc.invalid {
			assert.False(t, tc.lww1.Contains(obj), fmt.Sprintf("expected merged set to NOT contain %v", obj))
		}
	}
}

func TestConcurrent(t *testing.T) {
	i := 0
	lww := NewVolatile()
	for ; i < 100; i++ {
		setClock(int64(i))
		lww.Add(fmt.Sprintf("%v", i))
	}

	go func() {
		binary.Marshal(lww)
	}()

	var start, stop sync.WaitGroup
	start.Add(1)

	for x := 2; x < 10; x++ {
		other := NewVolatile()
		gi := i
		gu := x * 100

		for ; gi < gu; gi++ {
			setClock(int64(100000 + gi))
			other.Remove(fmt.Sprintf("%v", i))
		}

		stop.Add(1)
		go func() {
			start.Wait()
			lww.Merge(other)
			other.Merge(lww)
			stop.Done()
		}()
	}
	start.Done()
	stop.Wait()
}

// ------------------------------------------------------------------------------------

func TestRange(t *testing.T) {
	state := newVolatileWith(
		map[string]Time{
			"AC": {AddTime: 60, DelTime: 50},
			"AB": {AddTime: 60, DelTime: 50},
			"AA": {AddTime: 10, DelTime: 50}, // Deleted
			"BA": {AddTime: 60, DelTime: 50},
			"BB": {AddTime: 60, DelTime: 50},
			"BC": {AddTime: 60, DelTime: 50},
		})

	var count int
	state.Range([]byte("A"), func(_ string, v Time) bool {
		if v.IsAdded() {
			count++
		}
		return true
	})
	assert.Equal(t, 2, count)

	count = 0
	state.Range(nil, func(_ string, v Time) bool {
		if v.IsAdded() {
			count++
		}
		return true
	})
	assert.Equal(t, 5, count)
	assert.Equal(t, 6, state.Count())
}

// ------------------------------------------------------------------------------------

func TestTempMarshal(t *testing.T) {
	defer restoreClock(Now)

	setClock(0)
	state := newVolatileWith(map[string]Time{"A": {AddTime: 10, DelTime: 50}})

	// Encode
	enc, err := binary.Marshal(state)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x1, 0x1, 0x41, 0x2, 0x14, 0x64}, enc)

	// Decode
	dec := NewVolatile()
	err = binary.Unmarshal(enc, dec)
	assert.NoError(t, err)
	assert.Equal(t, state, dec)
}
