package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type secretStoreMock struct {
	mock.Mock
}

func (m *secretStoreMock) GetSecret(secretName string) (string, bool) {
	mockArgs := m.Called(secretName)
	v := mockArgs.Get(0).(string)
	return v, v != ""
}

func (m *secretStoreMock) Configure(c *Config) error {
	return nil
}

func Test_write(t *testing.T) {
	c := &Config{
		ListenAddr: ":80",
	}

	o := bytes.NewBuffer([]byte{})
	c.write(o)
	assert.Equal(t, "{\n\t\"listen\": \":80\",\n\t\"license\": \"\"\n}", string(o.Bytes()))
}

func Test_declassify(t *testing.T) {
	c := NewDefault()
	c.Vault = new(VaultConfig)
	m := new(secretStoreMock)
	m.On("GetSecret", "emitter/listen").Return(":999")
	m.On("GetSecret", "emitter/vault/address").Return("hello")
	m.On("GetSecret", mock.Anything).Return("")

	expected := NewDefault()
	expected.ListenAddr = ":999"
	expected.Vault = new(VaultConfig)
	expected.Vault.Address = "hello"

	c.declassify("emitter", m)

	assert.EqualValues(t, expected, c)

	c.Vault.Application = "abc"
	assert.True(t, c.HasVault())
}
