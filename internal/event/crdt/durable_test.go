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

func TestAddRemove(t *testing.T) {
	defer restoreClock(Now)
	for _, tc := range []struct {
		initial  Map
		expected Map
		actions  []Action
	}{
		{
			initial:  mapOf(false, T("A", 10, 0, "A1")),
			expected: mapOf(false, T("A", 20, 0, "A2")),
			actions:  []Action{T("A", 20, 0, "A2")},
		},
		{
			initial:  mapOf(false, T("A", 10, 0, "A1")),
			expected: mapOf(false, T("A", 10, 20, "A1")),
			actions:  []Action{T("A", 0, 20, "A1")},
		},
		{
			initial:  mapOf(false, T("A", 10, 0, "A1")),
			expected: mapOf(false, T("A", 20, 0, "A2")),
			actions:  []Action{T("A", 20, 0, "A2"), T("A", 15, 0, "A3")},
		},
		{
			initial:  mapOf(false, T("A", 10, 0, "A1")),
			expected: mapOf(false, T("A", 10, 20, "A1")),
			actions:  []Action{T("A", 0, 20), T("A", 0, 15)},
		},
		{
			initial:  mapOf(true, T("A", 10, 0, "A1")),
			expected: mapOf(true, T("A", 20, 0, "A2")),
			actions:  []Action{T("A", 20, 0, "A2")},
		},
		{
			initial:  mapOf(true, T("A", 10, 0, "A1")),
			expected: mapOf(true, T("A", 10, 20, "A1")),
			actions:  []Action{T("A", 0, 20, "A1")},
		},
		{
			initial:  mapOf(true, T("A", 10, 0, "A1")),
			expected: mapOf(true, T("A", 20, 0, "A2")),
			actions:  []Action{T("A", 20, 0, "A2"), T("A", 15, 0, "A3")},
		},
		{
			initial:  mapOf(true, T("A", 10, 0, "A1")),
			expected: mapOf(true, T("A", 10, 20, "A1")),
			actions:  []Action{T("A", 0, 20), T("A", 0, 15)},
		},
	} {
		for _, f := range tc.actions {
			k, v := f()
			if v.IsAdded() {
				setClock(v.AddTime())
				tc.initial.Add(k, v.Value())
			}
			if v.IsRemoved() {
				setClock(v.DelTime())
				tc.initial.Del(k)
			}

			equalSets(t, tc.expected, tc.initial)
			assert.Equal(t, tc.expected.Count(), tc.initial.Count())
		}
	}
}

func TestMerge(t *testing.T) {
	for _, tc := range []struct {
		lww1, expected Map
		lww2, delta    Map
		valid, invalid []string
	}{
		// Volatile -> Durable
		{
			lww1:     mapOf(true, T("A", 10, 0, "A1"), T("B", 20, 0, "B1")),
			lww2:     mapOf(false, T("A", 0, 20, "A2"), T("B", 0, 20, "B2")),
			expected: mapOf(true, T("A", 10, 20, "A2"), T("B", 20, 20, "B2")),
			delta:    mapOf(false, T("A", 0, 20, "A2"), T("B", 0, 20, "B2")),
			valid:    []string{"B"},
			invalid:  []string{"A"},
		},
		{
			lww1:     mapOf(true, T("A", 10, 0, "A1"), T("B", 20, 0, "B1")),
			lww2:     mapOf(false, T("A", 0, 20), T("B", 10, 0, "B2")),
			expected: mapOf(true, T("A", 10, 20), T("B", 20, 0, "B1")),
			delta:    mapOf(false, T("A", 0, 20)),
			valid:    []string{"B"},
			invalid:  []string{"A"},
		},
		{
			lww1:     mapOf(true, T("A", 30, 0, "A1"), T("B", 20, 0, "B1")),
			lww2:     mapOf(false, T("A", 20, 0, "A2"), T("B", 10, 0, "B2")),
			expected: mapOf(true, T("A", 30, 0, "A1"), T("B", 20, 0, "B1")),
			delta:    NewVolatile(),
			valid:    []string{"A", "B"},
			invalid:  []string{},
		},
		{
			lww1:     mapOf(true, T("A", 10, 0, "A1"), T("B", 0, 20)),
			lww2:     mapOf(false, T("C", 10, 0, "C1"), T("D", 0, 20)),
			expected: mapOf(true, T("A", 10, 0, "A1"), T("B", 0, 20), T("C", 10, 0, "C1"), T("D", 0, 20)),
			delta:    mapOf(false, T("C", 10, 0, "C1"), T("D", 0, 20)),
			valid:    []string{"A", "C"},
			invalid:  []string{"B", "D"},
		},
		{
			lww1:     mapOf(true, T("A", 10, 0, "A1"), T("B", 30, 0, "B1")),
			lww2:     mapOf(false, T("A", 20, 0, "A2"), T("B", 20, 0, "B2")),
			expected: mapOf(true, T("A", 20, 0, "A2"), T("B", 30, 0, "B1")),
			delta:    mapOf(false, T("A", 20, 0, "A2")),
			valid:    []string{"A", "B"},
			invalid:  []string{},
		},
		{
			lww1:     mapOf(true, T("A", 0, 10), T("B", 0, 30)),
			lww2:     mapOf(false, T("A", 0, 20), T("B", 0, 20)),
			expected: mapOf(true, T("A", 0, 20), T("B", 0, 30)),
			delta:    mapOf(false, T("A", 0, 20)),
			valid:    []string{},
			invalid:  []string{"A", "B"},
		},

		// Volatile -> Volatile
		{
			lww1:     mapOf(false, T("A", 10, 0, "A1"), T("B", 20, 0, "B1")),
			lww2:     mapOf(false, T("A", 0, 20, "A2"), T("B", 0, 20, "B2")),
			expected: mapOf(false, T("A", 10, 20, "A2"), T("B", 20, 20, "B2")),
			delta:    mapOf(false, T("A", 0, 20, "A2"), T("B", 0, 20, "B2")),
			valid:    []string{"B"},
			invalid:  []string{"A"},
		},
		{
			lww1:     mapOf(false, T("A", 10, 0, "A1"), T("B", 20, 0, "B1")),
			lww2:     mapOf(false, T("A", 0, 20), T("B", 10, 0, "B2")),
			expected: mapOf(false, T("A", 10, 20), T("B", 20, 0, "B1")),
			delta:    mapOf(false, T("A", 0, 20)),
			valid:    []string{"B"},
			invalid:  []string{"A"},
		},
		{
			lww1:     mapOf(false, T("A", 30, 0, "A1"), T("B", 20, 0, "B1")),
			lww2:     mapOf(false, T("A", 20, 0, "A2"), T("B", 10, 0, "B2")),
			expected: mapOf(false, T("A", 30, 0, "A1"), T("B", 20, 0, "B1")),
			delta:    NewVolatile(),
			valid:    []string{"A", "B"},
			invalid:  []string{},
		},
		{
			lww1:     mapOf(false, T("A", 10, 0, "A1"), T("B", 0, 20)),
			lww2:     mapOf(false, T("C", 10, 0, "C1"), T("D", 0, 20)),
			expected: mapOf(false, T("A", 10, 0, "A1"), T("B", 0, 20), T("C", 10, 0, "C1"), T("D", 0, 20)),
			delta:    mapOf(false, T("C", 10, 0, "C1"), T("D", 0, 20)),
			valid:    []string{"A", "C"},
			invalid:  []string{"B", "D"},
		},
		{
			lww1:     mapOf(false, T("A", 10, 0, "A1"), T("B", 30, 0, "B1")),
			lww2:     mapOf(false, T("A", 20, 0, "A2"), T("B", 20, 0, "B2")),
			expected: mapOf(false, T("A", 20, 0, "A2"), T("B", 30, 0, "B1")),
			delta:    mapOf(false, T("A", 20, 0, "A2")),
			valid:    []string{"A", "B"},
			invalid:  []string{},
		},
		{
			lww1:     mapOf(false, T("A", 0, 10), T("B", 0, 30)),
			lww2:     mapOf(false, T("A", 0, 20), T("B", 0, 20)),
			expected: mapOf(false, T("A", 0, 20), T("B", 0, 30)),
			delta:    mapOf(false, T("A", 0, 20)),
			valid:    []string{},
			invalid:  []string{"A", "B"},
		},
	} {

		tc.lww1.Merge(tc.lww2)
		equalSets(t, tc.expected, tc.lww1)
		equalSets(t, tc.delta, tc.lww2)

		for _, obj := range tc.valid {
			assert.True(t, tc.lww1.Has(obj), fmt.Sprintf("expected merged set to contain %v", obj))
		}

		for _, obj := range tc.invalid {
			assert.False(t, tc.lww1.Has(obj), fmt.Sprintf("expected merged set to NOT contain %v", obj))
		}
	}
}

func TestDurableConcurrent(t *testing.T) {
	i := 0
	lww := NewDurable("")
	defer lww.Close()
	for ; i < 100; i++ {
		setClock(int64(i))
		lww.Add(fmt.Sprintf("%v", i), nil)
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
			other.Del(fmt.Sprintf("%v", i))
		}

		stop.Add(1)
		go func() {
			start.Wait()
			lww.Merge(other)
			stop.Done()
		}()
	}
	start.Done()
	stop.Wait()
}

// ------------------------------------------------------------------------------------

func TestDurableRange(t *testing.T) {
	state := newDurableWith("",
		map[string]Value{
			"AC": newTime(60, 50, nil),
			"AB": newTime(60, 50, nil),
			"AA": newTime(10, 50, nil), // Deleted
			"BA": newTime(60, 50, nil),
			"BB": newTime(60, 50, nil),
			"BC": newTime(60, 50, nil),
		})

	var count int
	state.Range([]byte("A"), false, func(_ string, v Value) bool {
		count++
		return true
	})
	assert.Equal(t, 2, count)

	count = 0
	state.Range(nil, false, func(_ string, v Value) bool {
		count++
		return true
	})
	assert.Equal(t, 5, count)
}

// ------------------------------------------------------------------------------------

func TestDurableMarshal(t *testing.T) {
	defer restoreClock(Now)
	setClock(10)

	state := mapOf(true, T("A", 10, 50)).(*Durable)

	// Encode
	enc, err := binary.Marshal(state)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x1, 0x1, 0x41, 0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32}, enc)

	// Decode
	dec := NewDurable("")
	err = binary.Unmarshal(enc, dec)
	assert.NoError(t, err)
	println(dec.toMap()["A"].AddTime)
	assert.Equal(t, state.toMap(), dec.toMap())
}
