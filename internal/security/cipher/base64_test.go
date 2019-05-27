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

package cipher

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_decodeKey(t *testing.T) {
	// Just an error test since everything is already covered by other tests
	defer func(m [256]byte) { decodeMap = m }(decodeMap)
	for i := 0; i < len(decodeMap); i++ {
		decodeMap[i] = 1
	}

	in1 := []byte("#")
	_, err1 := decodeKey(make([]byte, 32), in1)
	assert.Error(t, err1)
	assert.NotEmpty(t, err1.Error())

	for i := 0; i < len(decodeMap); i++ {
		decodeMap[i] = byte(i)
	}

	for i := 0; i < 255; i++ {
		assert.NotPanics(t, func() {
			in2 := []byte{0, byte(i)}
			decodeKey(make([]byte, 32), in2)
		})
	}

}

func BenchmarkBase64(b *testing.B) {
	v := []byte("0TJnt4yZPL73zt35h1UTIFsYBLetyD_g")
	o := make([]byte, 24)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decodeKey(o, v)
	}
}

func TestBase64Decode(t *testing.T) {
	v := []byte("0TJnt4yZPL73zt35h1UTIFsYBLetyD_g")
	o1 := make([]byte, 24)
	o2 := make([]byte, 24)

	n1, e1 := decodeKey(o1, v)
	n2, e2 := base64.RawURLEncoding.Decode(o2, v)

	assert.Equal(t, o1, o2)
	assert.Equal(t, n1, n2)
	assert.Equal(t, e1, e2)
}

func TestBase64SelfDecode(t *testing.T) {
	v := []byte("0TJnt4yZPL73zt35h1UTIFsYBLetyD_g")
	o1 := []byte("0TJnt4yZPL73zt35h1UTIFsYBLetyD_g")
	o2 := make([]byte, 24)

	n1, e1 := decodeKey(o1, o1)
	n2, e2 := base64.RawURLEncoding.Decode(o2, v)

	assert.Equal(t, o1[:24], o2)
	assert.Equal(t, n1, n2)
	assert.Equal(t, e1, e2)
}
