package address

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	assert.Equal(t, uint64(0x313233343536), encode([]byte("123456")))

}
