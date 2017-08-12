package perf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSample(t *testing.T) {
	assert.Equal(t, 0, Sample().MemoryPrivate)

}
