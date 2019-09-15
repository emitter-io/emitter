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

package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKeyIsEmpty(t *testing.T) {
	key := Key([]byte{})
	assert.True(t, true, key.IsEmpty())
}

// BenchmarkValidateChannel-8   	10000000	       126 ns/op	      48 B/op	       1 allocs/op
func BenchmarkValidateChannel(b *testing.B) {
	key := Key(make([]byte, 24))
	key.SetTarget("a/")
	v := ParseChannel([]byte(string(key) + "a/b/c/d/"))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key.ValidateChannel(v)
	}
}

func validateChannel(k Key, c string) bool {
	return k.ValidateChannel(&Channel{Channel: []byte(c)})
}

func TestKey_New(t *testing.T) {
	key := Key(make([]byte, 24))

	// Test retro-compatibility
	// A key with bytes 12-13-14 set to 0 will only compare the first part of the channel.
	key.SetTarget("a/")
	key[12] = 0
	key[13] = 0
	key[14] = 0
	assert.True(t, key.ValidateChannel(ParseChannel([]byte(string(key)+"/a/"))))
	assert.True(t, key.ValidateChannel(ParseChannel([]byte(string(key)+"/a/b/"))))
	assert.True(t, key.ValidateChannel(ParseChannel([]byte(string(key)+"/a/b/c/"))))
	assert.True(t, key.ValidateChannel(ParseChannel([]byte(string(key)+"/a/+/c/"))))
	assert.False(t, key.ValidateChannel(ParseChannel([]byte(string(key)+"/b/"))))

	// Test exact channel
	key.SetTarget("a/b/c/")
	assert.False(t, validateChannel(key, "a/b/"))
	assert.True(t, validateChannel(key, "a/b/c/"))
	assert.False(t, validateChannel(key, "a/b/c/d/"))

	// Test exact channel with wildcard
	key.SetTarget("a/+/c/")
	assert.True(t, validateChannel(key, "a/b/c/"))
	assert.True(t, validateChannel(key, "a/c/c/"))
	assert.True(t, validateChannel(key, "a/d/c/"))
	assert.True(t, validateChannel(key, "a/+/c/"))
	assert.False(t, validateChannel(key, "a/b/+/"))

	key.SetTarget("+/")
	assert.True(t, validateChannel(key, "/"))
	assert.True(t, validateChannel(key, "a/"))
	assert.False(t, validateChannel(key, "a/b/"))
	assert.False(t, validateChannel(key, "a/b/c/"))

	key.SetTarget("+/+/")
	assert.False(t, validateChannel(key, "/"))
	assert.False(t, validateChannel(key, "a/"))
	assert.True(t, validateChannel(key, "a/b/"))
	assert.False(t, validateChannel(key, "a/b/c/"))
	assert.True(t, validateChannel(key, "a/+/"))
	assert.True(t, validateChannel(key, "+/b/"))
	assert.True(t, validateChannel(key, "+/+/"))

	key.SetTarget("+/+/+/")
	assert.False(t, validateChannel(key, "/"))
	assert.False(t, validateChannel(key, "a/"))
	assert.False(t, validateChannel(key, "a/b/"))
	assert.True(t, validateChannel(key, "a/b/c/"))

	// Test open channel
	key.SetTarget("#/")
	assert.True(t, validateChannel(key, "/"))
	assert.True(t, validateChannel(key, "a/"))
	assert.True(t, validateChannel(key, "a/b/"))
	assert.True(t, validateChannel(key, "a/b/c/"))

	key.SetTarget("a/b/c/#/")
	assert.False(t, validateChannel(key, "a/b/"))
	assert.True(t, validateChannel(key, "a/b/c/"))
	assert.True(t, validateChannel(key, "a/b/c/d/"))
	assert.True(t, validateChannel(key, "a/b/c/d/e/"))
	assert.True(t, validateChannel(key, "a/b/c/d/+/f/"))
	assert.True(t, validateChannel(key, "a/b/c/d/+/f/#/"))

	// Test ErrTargetTooLong
	assert.Nil(t, key.SetTarget("1/2/3/4/5/6/7/8/9/10/11/12/13/14/15/16/17/18/19/20/21/22/23/"))
	assert.Nil(t, key.SetTarget("1/2/3/4/5/6/7/8/9/10/11/12/13/14/15/16/17/18/19/20/21/22/23/#/"))
	assert.Equal(t, ErrTargetTooLong, key.SetTarget("1/2/3/4/5/6/7/8/9/10/11/12/13/14/15/16/17/18/19/20/21/22/23/24/"))
}

func TestKey(t *testing.T) {
	key := Key(make([]byte, 24))

	key.SetSalt(999)
	key.SetMaster(2)
	key.SetContract(123)
	key.SetSignature(777)
	key.SetPermissions(AllowReadWrite)
	key.SetTarget("a/b/c/")
	key.SetExpires(time.Unix(1497683272, 0).UTC())

	assert.Equal(t, uint16(999), key.Salt())
	assert.Equal(t, uint16(2), key.Master())
	assert.Equal(t, uint32(123), key.Contract())
	assert.Equal(t, uint32(777), key.Signature())
	assert.Equal(t, AllowReadWrite, key.Permissions())
	assert.Equal(t, time.Unix(1497683272, 0).UTC(), key.Expires())

	key.SetExpires(time.Unix(0, 0))
	assert.Equal(t, time.Unix(0, 0).UTC(), key.Expires())
	assert.False(t, key.IsExpired())

	key.SetExpires(time.Unix(1497683272, 0).UTC())
	assert.True(t, key.IsExpired())

	assert.True(t, key.HasPermission(AllowRead))
	assert.False(t, key.HasPermission(AllowExtend))

	key.SetPermission(AllowExtend, true)
	assert.True(t, key.HasPermission(AllowExtend))

	key.SetPermission(AllowExtend, false)
	assert.False(t, key.HasPermission(AllowExtend))

	key.SetPermissions(AllowMaster)
	assert.True(t, key.IsMaster())
	assert.True(t, key.HasPermission(AllowMaster))
}
