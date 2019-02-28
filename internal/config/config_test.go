package config

import (
	"os"
	"strings"
	"testing"

	"github.com/emitter-io/config/dynamo"
	"github.com/stretchr/testify/assert"
)

func Test_NewDefaut(t *testing.T) {
	c := NewDefault().(*Config)
	assert.Equal(t, ":8080", c.ListenAddr)
	//assert.Nil(t, c.Vault())

	tls, _, ok := c.Certificate()
	assert.Nil(t, tls)
	assert.False(t, ok)
}

func Test_Addr(t *testing.T) {
	c := &Config{
		ListenAddr: "private",
	}

	addr := c.Addr()
	assert.True(t, strings.HasSuffix(addr.String(), ":8080"))
}

func Test_AddrInvalid(t *testing.T) {
	assert.Panics(t, func() {
		c := &Config{ListenAddr: "g3ew235wgs"}
		c.Addr()
	})
}

func Test_New(t *testing.T) {
	c := New("test.conf", dynamo.NewProvider())
	defer os.Remove("test.conf")

	assert.NotNil(t, c)
}

func Test_DefaultMaxMessageSize(t *testing.T){
	c := New("test.conf", dynamo.NewProvider())
	defer os.Remove("test.conf")

	assert.EqualValues(t, c.MaxMessageBytes(), maxMessageSize)
}
