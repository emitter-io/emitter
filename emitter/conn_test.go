package emitter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_nextIdentifier(t *testing.T) {
	now := uint64(time.Now().UTC().Unix())
	n1 := nextIdentifier()
	assert.Equal(t, true, n1 >= now)
	n2 := nextIdentifier()
	assert.Equal(t, true, n2 >= now)
	assert.Equal(t, true, n2 >= n1)
}
