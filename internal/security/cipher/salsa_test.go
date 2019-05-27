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
	"time"

	"github.com/emitter-io/emitter/internal/security"
	"github.com/stretchr/testify/assert"
)

type x struct {
	test string
}

// BenchmarkEncryptKey2-8   	 5000000	       356 ns/op	      64 B/op	       2 allocs/op
func Benchmark_Salsa_EncryptKey(b *testing.B) {
	cipher := new(Salsa)
	key := "um4m30suos9k0tNjZiO19FyGNtmZjRlN"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cipher.EncryptKey([]byte(key))
	}
}

func Test_Salsa(t *testing.T) {
	cipher := new(Salsa)
	key := security.Key(make([]byte, 24))
	key.SetSalt(999)
	key.SetMaster(2)
	key.SetContract(123)
	key.SetSignature(777)
	key.SetPermissions(security.AllowReadWrite)
	key.SetTarget("a/b/c/")
	key.SetExpires(time.Unix(1497683272, 0).UTC())

	encoded, err := cipher.EncryptKey(key)
	assert.NoError(t, err)
	assert.Equal(t, "uYkm3UsuorRk0tBqliO18gs5xXmXioMF", encoded)

	decoded, err := cipher.DecryptKey([]byte(encoded))
	assert.NoError(t, err)
	assert.Equal(t, key, decoded)
}

// Benchmark_Salsa_DecryptKey-8   	 5000000	       309 ns/op	       0 B/op	       0 allocs/op
func Benchmark_Salsa_DecryptKey(b *testing.B) {
	cipher := new(Salsa)
	key := "um4m30suos9k0tNjZiO19FyGNtmZjRlN"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cipher.DecryptKey([]byte(key))
	}
}

func Test_Salsa_Errors(t *testing.T) {
	cipher := new(Salsa)
	tests := []struct {
		key string
		err bool
	}{

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
		_, err := cipher.DecryptKey([]byte(tc.key))
		assert.Equal(t, tc.err, err != nil, tc.key)

	}
}

func TestNewSalsa(t *testing.T) {

	// Happy path
	{
		c, err := NewSalsa(make([]byte, 32), make([]byte, 24))
		assert.NoError(t, err)
		assert.NotNil(t, c)
	}

	// Error case
	{
		c, err := NewSalsa(nil, nil)
		assert.Error(t, err)
		assert.Nil(t, c)
	}

}
