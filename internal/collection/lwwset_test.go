/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
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

package collection

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLWWESetAddContains(t *testing.T) {
	testStr := "ABCD"

	lww := NewLWWSet()
	assert.False(t, lww.Contains(testStr))

	lww.Add(testStr)
	assert.True(t, lww.Contains(testStr))

	entry := lww.Set[testStr]
	assert.True(t, entry.IsAdded())
	assert.False(t, entry.IsRemoved())
	assert.False(t, entry.IsZero())
}

func TestLWWESetAddRemoveContains(t *testing.T) {
	lww := NewLWWSet()
	testStr := "object2"

	lww.Add(testStr)
	time.Sleep(1 * time.Millisecond)
	lww.Remove(testStr)

	assert.False(t, lww.Contains(testStr))

	entry := lww.Set[testStr]
	assert.False(t, entry.IsAdded())
	assert.True(t, entry.IsRemoved())
	assert.False(t, entry.IsZero())
}

func TestLWWESetMerge(t *testing.T) {
	var T = func(add, del int64) LWWTime {
		return LWWTime{AddTime: add, DelTime: del}
	}

	for _, tc := range []struct {
		lww1, lww2, expected, delta *LWWSet
		valid, invalid              []string
	}{
		{
			lww1: &LWWSet{
				Set: LWWState{"A": T(10, 0), "B": T(20, 0)},
			},
			lww2: &LWWSet{
				Set: LWWState{"A": T(0, 20), "B": T(0, 20)},
			},
			expected: &LWWSet{
				Set: LWWState{"A": T(10, 20), "B": T(20, 20)},
			},
			delta: &LWWSet{
				Set: LWWState{"A": T(0, 20), "B": T(0, 20)},
			},
			valid:   []string{"B"},
			invalid: []string{"A"},
		},
		{
			lww1: &LWWSet{
				Set: LWWState{"A": T(10, 0), "B": T(20, 0)},
			},
			lww2: &LWWSet{
				Set: LWWState{"A": T(0, 20), "B": T(10, 0)},
			},
			expected: &LWWSet{
				Set: LWWState{"A": T(10, 20), "B": T(20, 0)},
			},
			delta: &LWWSet{
				Set: LWWState{"A": T(0, 20)},
			},
			valid:   []string{"B"},
			invalid: []string{"A"},
		},
		{
			lww1: &LWWSet{
				Set: LWWState{"A": T(30, 0), "B": T(20, 0)},
			},
			lww2: &LWWSet{
				Set: LWWState{"A": T(20, 0), "B": T(10, 0)},
			},
			expected: &LWWSet{
				Set: LWWState{"A": T(30, 0), "B": T(20, 0)},
			},
			delta: &LWWSet{
				Set: LWWState{},
			},
			valid:   []string{"A", "B"},
			invalid: []string{},
		},
		{
			lww1: &LWWSet{
				Set: LWWState{"A": T(10, 0), "B": T(0, 20)},
			},
			lww2: &LWWSet{
				Set: LWWState{"C": T(10, 0), "D": T(0, 20)},
			},
			expected: &LWWSet{
				Set: LWWState{"A": T(10, 0), "B": T(0, 20), "C": T(10, 0), "D": T(0, 20)},
			},
			delta: &LWWSet{
				Set: LWWState{"C": T(10, 0), "D": T(0, 20)},
			},
			valid:   []string{"A", "C"},
			invalid: []string{"B", "D"},
		},
		{
			lww1: &LWWSet{
				Set: LWWState{"A": T(10, 0), "B": T(30, 0)},
			},
			lww2: &LWWSet{
				Set: LWWState{"A": T(20, 0), "B": T(20, 0)},
			},
			expected: &LWWSet{
				Set: LWWState{"A": T(20, 0), "B": T(30, 0)},
			},
			delta: &LWWSet{
				Set: LWWState{"A": T(20, 0)},
			},
			valid:   []string{"A", "B"},
			invalid: []string{},
		},
		{
			lww1: &LWWSet{
				Set: LWWState{"A": T(0, 10), "B": T(0, 30)},
			},
			lww2: &LWWSet{
				Set: LWWState{"A": T(0, 20), "B": T(0, 20)},
			},
			expected: &LWWSet{
				Set: LWWState{"A": T(0, 20), "B": T(0, 30)},
			},
			delta: &LWWSet{
				Set: LWWState{"A": T(0, 20)},
			},
			valid:   []string{},
			invalid: []string{"A", "B"},
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
	lww := NewLWWSet()
	lww.Add("A")
	lww.Add("B")
	lww.Add("C")

	all := lww.All()
	assert.Equal(t, 3, len(all))
}

func TestLWWESetGC(t *testing.T) {
	defer restoreClock(Now)

	setClock(0)
	lww := NewLWWSet()
	lww.Add("A")
	lww.Add("B")
	lww.Add("C")

	setClock(1)
	lww.Remove("B")
	lww.Remove("C")

	setClock(gcCutoff + 2)

	lww.GC()
	assert.Equal(t, 1, len(lww.Set))
}

// RestoreClock restores the clock time
func restoreClock(clk clock) {
	Now = clk
}

// SetClock sets the clock time for testing
func setClock(t int64) {
	Now = func() int64 { return t }
	println("clock set to", Now())
}
