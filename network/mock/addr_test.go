package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddr(t *testing.T) {
	addr := Addr{
		NetworkString: "1",
		AddrString:    "2",
	}

	assert.Equal(t, "1", addr.Network())
	assert.Equal(t, "2", addr.String())
}
