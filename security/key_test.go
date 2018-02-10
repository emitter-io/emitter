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

func validateChannel(k Key, c string) bool {
	return k.ValidateChannel(&Channel{Channel: []byte(c)})
}

func TestKey_New(t *testing.T) {
	key := Key(make([]byte, 24))

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

	// Test open channel
	key.SetTarget("a/b/c/#/")
	assert.False(t, validateChannel(key, "a/b/"))
	assert.True(t, validateChannel(key, "a/b/c/"))
	assert.True(t, validateChannel(key, "a/b/c/d/"))
	assert.True(t, validateChannel(key, "a/b/c/d/e/"))
	assert.True(t, validateChannel(key, "a/b/c/d/+/f/"))
	assert.True(t, validateChannel(key, "a/b/c/d/+/f/#/"))

	assert.Nil(t, key.SetTarget("1/2/3/4/5/6/7/8/9/10/11/12/13/14/15/"))
	assert.Nil(t, key.SetTarget("1/2/3/4/5/6/7/8/9/10/11/12/13/14/15/#/"))
	assert.NotNil(t, key.SetTarget("1/2/3/4/5/6/7/8/9/10/11/12/13/14/15/16/"))
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

	key.SetPermissions(AllowMaster)
	assert.True(t, key.IsMaster())
	assert.True(t, key.HasPermission(AllowMaster))
}
