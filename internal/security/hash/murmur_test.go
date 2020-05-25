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

package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// BenchmarkOf-8   	99755598	        12.1 ns/op	       0 B/op	       0 allocs/op
func BenchmarkOf(b *testing.B) {
	v := []byte("a/b/c/d/e/f/g/h/this/is/emitter")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Of(v)
	}
}

// BenchmarkOfString-8   	79873267	        15.0 ns/op	       0 B/op	       0 allocs/op
func BenchmarkOfString(b *testing.B) {
	v := "a/b/c/d/e/f/g/h/this/is/emitter"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = OfString(v)
	}
}

func TestKeyBanHash(t *testing.T) {
	h := OfString("keyban")
	assert.Equal(t, uint32(861724010), h)
}

func TestMeHash(t *testing.T) {
	h := OfString("me")
	assert.Equal(t, uint32(2539734036), h)
}

func TestShareHash(t *testing.T) {
	h := Of([]byte("$share"))
	assert.Equal(t, uint32(1480642916), h)
}

func TestLinkHash(t *testing.T) {
	h := Of([]byte("link"))
	assert.Equal(t, uint32(2667034312), h)
}

func TestGetHash(t *testing.T) {
	h := Of([]byte("+"))
	if h != 1815237614 {
		t.Errorf("Hash %d is not equal to %d", h, 1815237614)
	}
}

func TestGetHash2(t *testing.T) {
	h := Of([]byte("hello world"))
	if h != 4008393376 {
		t.Errorf("Hash %d is not equal to %d", h, 1815237614)
	}
}
