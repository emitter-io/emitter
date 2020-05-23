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
	"time"

	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
)

func TestLWWESetAddContains(t *testing.T) {
	testStr := "ABCD"

	lww := New()
	assert.False(t, lww.Contains(testStr))

	lww.Add(testStr)
	assert.True(t, lww.Contains(testStr))

	entry := lww.data[testStr]
	assert.True(t, entry.IsAdded())
	assert.False(t, entry.IsRemoved())
	assert.False(t, entry.IsZero())
}

func TestLWWESetAddRemoveContains(t *testing.T) {
	lww := New()
	testStr := "object2"

	lww.Add(testStr)
	time.Sleep(1 * time.Millisecond)
	lww.Remove(testStr)

	assert.False(t, lww.Contains(testStr))

	entry := lww.data[testStr]
	assert.False(t, entry.IsAdded())
	assert.True(t, entry.IsRemoved())
	assert.False(t, entry.IsZero())
}

func TestLWWESetMerge(t *testing.T) {
	var T = func(add, del int64) Time {
		return Time{AddTime: add, DelTime: del}
	}

	for _, tc := range []struct {
		lww1, lww2, expected, delta *Set
		valid, invalid              []string
	}{
		{
			lww1:     NewWith(map[string]Time{"A": T(10, 0), "B": T(20, 0)}),
			lww2:     NewWith(map[string]Time{"A": T(0, 20), "B": T(0, 20)}),
			expected: NewWith(map[string]Time{"A": T(10, 20), "B": T(20, 20)}),
			delta:    NewWith(map[string]Time{"A": T(0, 20), "B": T(0, 20)}),
			valid:    []string{"B"},
			invalid:  []string{"A"},
		},
		{
			lww1:     NewWith(map[string]Time{"A": T(10, 0), "B": T(20, 0)}),
			lww2:     NewWith(map[string]Time{"A": T(0, 20), "B": T(10, 0)}),
			expected: NewWith(map[string]Time{"A": T(10, 20), "B": T(20, 0)}),
			delta:    NewWith(map[string]Time{"A": T(0, 20)}),
			valid:    []string{"B"},
			invalid:  []string{"A"},
		},
		{
			lww1:     NewWith(map[string]Time{"A": T(30, 0), "B": T(20, 0)}),
			lww2:     NewWith(map[string]Time{"A": T(20, 0), "B": T(10, 0)}),
			expected: NewWith(map[string]Time{"A": T(30, 0), "B": T(20, 0)}),
			delta:    New(),
			valid:    []string{"A", "B"},
			invalid:  []string{},
		},
		{
			lww1:     NewWith(map[string]Time{"A": T(10, 0), "B": T(0, 20)}),
			lww2:     NewWith(map[string]Time{"C": T(10, 0), "D": T(0, 20)}),
			expected: NewWith(map[string]Time{"A": T(10, 0), "B": T(0, 20), "C": T(10, 0), "D": T(0, 20)}),
			delta:    NewWith(map[string]Time{"C": T(10, 0), "D": T(0, 20)}),
			valid:    []string{"A", "C"},
			invalid:  []string{"B", "D"},
		},
		{
			lww1:     NewWith(map[string]Time{"A": T(10, 0), "B": T(30, 0)}),
			lww2:     NewWith(map[string]Time{"A": T(20, 0), "B": T(20, 0)}),
			expected: NewWith(map[string]Time{"A": T(20, 0), "B": T(30, 0)}),
			delta:    NewWith(map[string]Time{"A": T(20, 0)}),
			valid:    []string{"A", "B"},
			invalid:  []string{},
		},
		{
			lww1:     NewWith(map[string]Time{"A": T(0, 10), "B": T(0, 30)}),
			lww2:     NewWith(map[string]Time{"A": T(0, 20), "B": T(0, 20)}),
			expected: NewWith(map[string]Time{"A": T(0, 20), "B": T(0, 30)}),
			delta:    NewWith(map[string]Time{"A": T(0, 20)}),
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

func TestLWWESetAll(t *testing.T) {
	defer restoreClock(Now)

	setClock(0)
	lww := New()
	lww.Add("A")
	lww.Add("B")
	lww.Add("C")

	all := lww.Clone()
	assert.Equal(t, 3, len(all.data))
}

func TestLWWESetGC(t *testing.T) {
	defer restoreClock(Now)

	setClock(0)
	lww := New()
	lww.Add("A")
	lww.Add("B")
	lww.Add("C")

	setClock(1)
	lww.Remove("B")
	lww.Remove("C")

	setClock(gcCutoff + 2)

	lww.GC()
	assert.Equal(t, 1, len(lww.data))
}

func TestConcurrent(t *testing.T) {
	i := 0
	lww := New()
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
		other := New()
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

// ------------------------------------------------------------------------------------

func TestRange(t *testing.T) {
	state := NewWith(
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
}
