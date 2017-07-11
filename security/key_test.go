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

func TestKey(t *testing.T) {
	key := Key(make([]byte, 24))

	key.SetSalt(999)
	key.SetMaster(2)
	key.SetContract(123)
	key.SetSignature(777)
	key.SetPermissions(AllowReadWrite)
	key.SetTarget(56789)
	key.SetExpires(time.Unix(1497683272, 0).UTC())

	assert.Equal(t, uint16(999), key.Salt())
	assert.Equal(t, uint16(2), key.Master())
	assert.Equal(t, int32(123), key.Contract())
	assert.Equal(t, int32(777), key.Signature())
	assert.Equal(t, AllowReadWrite, key.Permissions())
	assert.Equal(t, uint32(56789), key.Target())
	assert.Equal(t, time.Unix(1497683272, 0).UTC(), key.Expires())

	key.SetExpires(time.Unix(0, 0))
	assert.Equal(t, time.Unix(0, 0).UTC(), key.Expires())
	assert.False(t, key.IsExpired())

	key.SetExpires(time.Unix(1497683272, 0).UTC())
	assert.True(t, key.IsExpired())

	key.SetPermissions(AllowMaster)
	assert.True(t, key.IsMaster())
}
