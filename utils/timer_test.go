package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRepeat(t *testing.T) {
	closing := make(chan bool)
	count := 0
	defer close(closing)

	Repeat(func() { count++ }, 1*time.Microsecond, closing)
	time.Sleep(50 * time.Millisecond)
	assert.True(t, count > 0)
}
