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

package message

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKeyCodec(t *testing.T) {
	m := Message{
		Ssid: Ssid{1, 2, 3, 4},
		Time: time.Now().UnixNano(),
	}

	key := m.Key()
	assert.NotEmpty(t, key)

	ok := key.Match(Ssid{1, 2, 3, 4}, 0, math.MaxInt64)
	fmt.Println(m, key)
	assert.True(t, ok)
}

func BenchmarkKeyEncode(b *testing.B) {
	m := Message{
		Ssid: Ssid{1, 2, 3},
		Time: time.Now().UnixNano(),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		m.Key()
	}
}

func BenchmarkKeyDecode(b *testing.B) {
	m := Message{
		Ssid: Ssid{1, 2, 3},
		Time: time.Now().UnixNano(),
	}
	key := m.Key()
	query := Ssid{1, 2, 3}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		key.Match(query, 0, math.MaxInt64)
	}
}

func BenchmarkKeyMatch(b *testing.B) {
	m := Message{
		Ssid: Ssid{1, 2, 3, 4, 5},
		Time: 1,
	}
	item := []byte(m.Key())
	ssid := Ssid{1, 2, 3, 4, 5}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Key(item).Match(ssid, 0, math.MaxInt64)
	}
}
