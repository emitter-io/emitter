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
	"crypto/rand"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/security"
	"github.com/stretchr/testify/assert"
)

type x struct {
	test string
}

// Benchmark_Shuffled_EncryptKey-8   	 4477147	       263 ns/op	      64 B/op	       2 allocs/op
func Benchmark_Shuffled_EncryptKey(b *testing.B) {
	cipher := new(Shuffle)
	key := "A-dOBQDuXhqoFz-GZZdbpSFCtzmFl7Ng"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cipher.EncryptKey([]byte(key))
	}
}

func Test_Shuffled(t *testing.T) {
	cipher := new(Shuffle)
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
	assert.Equal(t, "A-dOBQDuXhqoFz-GZZdbpSFCtzmFl7Ng", encoded)

	decoded, err := cipher.DecryptKey([]byte(encoded))
	assert.NoError(t, err)
	assert.Equal(t, key, decoded)
}

// Benchmark_Shuffled_DecryptKey-8   	 4857697	       245 ns/op	       0 B/op	       0 allocs/op
func Benchmark_Shuffled_DecryptKey(b *testing.B) {
	cipher := new(Shuffle)
	key := "A-dOBQDuXhqoFz-GZZdbpSFCtzmFl7Ng"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cipher.DecryptKey([]byte(key))
	}
}

func Test_Shuffled_Errors(t *testing.T) {
	cipher := new(Shuffle)
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

func TestNewShuffled(t *testing.T) {

	// Happy path
	{
		c, err := NewShuffle(make([]byte, 32), make([]byte, 16))
		assert.NoError(t, err)
		assert.NotNil(t, c)
	}

	// Error case
	{
		c, err := NewShuffle(nil, nil)
		assert.Error(t, err)
		assert.Nil(t, c)
	}

}

func TestShuffled_Entropy(t *testing.T) {
	cryptoKey := make([]byte, 32)
	rand.Read(cryptoKey)

	nonce := make([]byte, 16)
	rand.Read(nonce)

	c, err := NewShuffle(cryptoKey, nonce)

	key1 := makeKey(111)
	key2 := makeKey(333)

	k1, err := c.EncryptKey(key1)
	assert.NoError(t, err)

	k2, err := c.EncryptKey(key2)
	assert.NoError(t, err)

	var diff int
	for i := range k1 {
		if k1[i] != k2[i] {
			diff++
		}
	}

	assert.NotEqual(t, k1, k2)
	assert.Greater(t, diff, 20)
}

func printKey(key security.Key) {
	println(key.Salt(), key.Contract(), key.Signature(), key.Expires().String())
}

func makeKey(salt int) security.Key {
	key := security.Key(make([]byte, 24))
	key.SetSalt(uint16(salt))
	key.SetMaster(2)
	key.SetContract(123)
	key.SetSignature(777)
	key.SetPermissions(security.AllowReadWrite)
	key.SetTarget("a/b/c/")
	key.SetExpires(time.Unix(1497683272, 0).UTC())
	return key
}
