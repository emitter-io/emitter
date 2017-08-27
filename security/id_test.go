package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	defer func(n uint64) { next = n }(next)

	next = 0
	i1 := NewID()
	i2 := NewID()

	assert.Equal(t, ID(1), i1)
	assert.Equal(t, ID(2), i2)
}

func TestIDToString(t *testing.T) {
	defer func(n uint64) { next = n }(next)

	next = 0
	i1 := NewID()
	i2 := NewID()

	assert.Equal(t, "01", i1.String())
	assert.Equal(t, "02", i2.String())
}

func TestIDToUnique(t *testing.T) {
	defer func(n uint64) { next = n }(next)

	next = 0
	i1 := NewID()
	i2 := NewID()

	assert.Equal(t, "F45JPXDSXVRWBUKTDNCCM4PGQI", i1.Unique(123, "hello"))
	assert.Equal(t, "XCFU2OA7OO2COPZOJ5VA6GS6BM", i2.Unique(123, "hello"))
}
