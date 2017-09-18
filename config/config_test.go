package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewDefaut(t *testing.T) {
	c := NewDefault().(*Config)
	assert.Equal(t, ":8080", c.ListenAddr)
	assert.Nil(t, c.Vault())
}
