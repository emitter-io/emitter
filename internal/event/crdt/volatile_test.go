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

func TestVolatile_Concurrent(t *testing.T) {
	i := 0
	lww := NewVolatile()
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
	assert.Equal(t, 6, state.Count())
}

// ------------------------------------------------------------------------------------

func TestTempMarshal(t *testing.T) {
	defer restoreClock(Now)

	setClock(0)
	state := newVolatileWith(map[string]Value{"A": newTime(10, 50, nil)})

	// Encode
	enc, err := binary.Marshal(state)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x1, 0x1, 0x41, 0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32}, enc)

	// Decode
	dec := NewVolatile()
	err = binary.Unmarshal(enc, dec)
	assert.NoError(t, err)
	assert.Equal(t, state, dec)
}
