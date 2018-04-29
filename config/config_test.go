package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewDefaut(t *testing.T) {
	c := NewDefault().(*Config)
	assert.Equal(t, ":8080", c.ListenAddr)
	assert.Nil(t, c.Vault())

	tls, _, ok := c.Certificate()
	assert.Nil(t, tls)
	assert.False(t, ok)
}
