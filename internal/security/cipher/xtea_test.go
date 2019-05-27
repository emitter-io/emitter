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
	"testing"

	"github.com/stretchr/testify/assert"
)

// BenchmarkEncryptKey-8   	 3000000	       402 ns/op	      64 B/op	       2 allocs/op
func Benchmark_Xtea_EncryptKey(b *testing.B) {
	cipher := &Xtea{key: [4]uint32{3443472288, 896798054, 972856492, 1831128908}}
	key := "0TJnt4yZPL73zt35h1UTIFsYBLetyD_g"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cipher.EncryptKey([]byte(key))
	}
}

// BenchmarkDecryptKey-8   	 3000000	       385 ns/op	       0 B/op	       0 allocs/op
func Benchmark_Xtea_DecryptKey(b *testing.B) {
	cipher := &Xtea{key: [4]uint32{3443472288, 896798054, 972856492, 1831128908}}
	key := "0TJnt4yZPL73zt35h1UTIFsYBLetyD_g"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cipher.DecryptKey([]byte(key))
	}
}

func Test_Xtea_EncryptDecrypt(t *testing.T) {

	// license: zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI
	// secret key: kBCZch5re3Ue-kpG1Aa8Vo7BYvXZ3UwR
	cipher := &Xtea{key: [4]uint32{3443472288, 896798054, 972856492, 1831128908}}
	tests := []struct {
		key     string
		channel string
		acl     string
		err     bool
	}{
		{
			key:     "0TJnt4yZPL73zt35h1UTIFsYBLetyD_g",
			channel: "emitter",
			acl:     "rsp",
		},
		{
			key:     "xm54Sj0srWlSEctra-yU6ZA6Z2e6pp7c",
			channel: "a/b/c",
			acl:     "rw",
		},
		{
			key:     "YXiszDIuAbkiyJfG-J1YpAwI0jUAbXW_",
			channel: "a/b/c",
			acl:     "sl",
		},
		{
			key:     "-ZgVnx1gr7BxxRCDsrEXBvmLVz86vGzs",
			channel: "cluster/#/",
			acl:     "rwsl",
		},
		{
			key:     "EbUlduEbUssgWueAWjkEZwdYG5YC0dGh",
			channel: "a/b/c/",
			acl:     "rwslpex",
		},
		{
			key: "",
			err: true,
		},
		{
			key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa*",
			err: true,
		},
	}

	for _, tc := range tests {
		key, err := cipher.DecryptKey([]byte(tc.key))
		assert.Equal(t, tc.err, err != nil, tc.key)
		if tc.err {
			continue
		}

		assert.Equal(t, uint32(989603869), key.Contract())
		encrypted, err := cipher.EncryptKey(key)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, tc.key, encrypted)
	}
}

func Test_Xtea_encrypt(t *testing.T) {
	// Just an error test since everything is already covered by other tests
	cipher := &Xtea{key: [4]uint32{3443472288, 896798054, 972856492, 1831128908}}
	err := cipher.encrypt(make([]byte, 10))
	assert.Error(t, err)
}

func TestNewXtea(t *testing.T) {
	tests := []struct {
		key      string
		expected [4]uint32
		err      bool
	}{
		{key: "zT83oDV0DWY5_JysbSTPTA", expected: [4]uint32{3443472288, 896798054, 972856492, 1831128908}},
		{key: "#%#%^", err: true},
		{key: "aaa", err: true},
	}

	for _, tc := range tests {
		cipher, err := NewXtea(tc.key)
		assert.Equal(t, tc.err, err != nil)
		if !tc.err {
			assert.EqualValues(t, tc.expected, cipher.key)
		}
	}
}
