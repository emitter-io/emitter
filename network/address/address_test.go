package address

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	assert.Equal(t, "MTIzNDU2", encode([]byte("123456")))
}
