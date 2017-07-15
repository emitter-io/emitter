package broker

import (
	"testing"

	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/network/address"
	"github.com/stretchr/testify/assert"
)

func TestService_clusterConfig(t *testing.T) {
	cfg := config.NewDefault()
	s := new(Service)

	o := s.clusterConfig(cfg)
	assert.Equal(t, o.MemberlistConfig.SecretKey, cfg.Cluster.Key())
	assert.Equal(t, o.MemberlistConfig.AdvertiseAddr, address.External().String())
	assert.Equal(t, o.MemberlistConfig.BindPort, cfg.Cluster.Port)
}

func TestService_Name(t *testing.T) {
	s := new(Service)
	assert.Equal(t, "", s.Name())

	cfg := config.NewDefault()
	cfg.Cluster.NodeName = "hello"
	_ = s.clusterConfig(cfg)
	assert.Equal(t, "hello", s.Name())
}
