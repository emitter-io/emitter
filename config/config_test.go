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
		TCPPort: ":80",
		TLSPort: ":443",
	}

	o := bytes.NewBuffer([]byte{})
	c.write(o)
	assert.Equal(t, "{\n\t\"tcp\": \":80\",\n\t\"tls\": \":443\",\n\t\"license\": \"\"\n}", string(o.Bytes()))
}

func TestClusterKey(t *testing.T) {
	c := ClusterConfig{}

	assert.Nil(t, c.Key())

	c.ClusterKey = "hi"
	key := c.Key()

	assert.True(t, true, len(key) == 16)
	assert.Equal(t, []byte{0x91, 0x3c, 0xca, 0x63, 0xe1, 0x36, 0x9, 0xc7, 0x86, 0x59, 0xa2, 0xd2, 0x16, 0x4, 0x50, 0xf1}, key)
}

func Test_declassify(t *testing.T) {
	c := NewDefault()
	c.Vault = new(VaultConfig)
	m := new(secretStoreMock)
	m.On("GetSecret", "emitter/tcp").Return(":999")
	m.On("GetSecret", "emitter/cluster/gossip").Return("123")
	m.On("GetSecret", "emitter/vault/address").Return("hello")
	m.On("GetSecret", mock.Anything).Return("")

	expected := NewDefault()
	expected.TCPPort = ":999"
	expected.Vault = new(VaultConfig)
	expected.Vault.Address = "hello"
	expected.Cluster.Gossip = 123

	c.declassify("emitter", m)

	assert.EqualValues(t, expected, c)

	c.Vault.Application = "abc"
	assert.True(t, c.HasVault())
}
