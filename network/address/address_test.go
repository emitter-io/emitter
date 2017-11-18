package address

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	assert.Equal(t, uint64(0x313233343536), encode([]byte("123456")))
}

func TestGlobals(t *testing.T) {
	assert.NotEqual(t, "", External().String())
	assert.NotEqual(t, Fingerprint(0), Hardware())
	assert.NotEqual(t, "", Hardware().String())
	assert.NotEqual(t, "", Hardware().Hex())
}
