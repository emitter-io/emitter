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

	Repeat(func() { count++ }, 1*time.Millisecond, closing)
	time.Sleep(10 * time.Millisecond)
	assert.True(t, count > 0)
}
