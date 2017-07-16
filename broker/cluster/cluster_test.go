package cluster

import (
	"testing"
	"time"

	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/network/address"
	"github.com/hashicorp/serf/serf"
	"github.com/stretchr/testify/assert"
)

func TestCluster_clusterConfig(t *testing.T) {
	cfg := config.NewDefault()
	s, err := NewCluster(cfg.Cluster, nil)
	assert.NoError(t, err)

	o := s.config
	assert.Equal(t, o.MemberlistConfig.SecretKey, cfg.Cluster.Key())
	assert.Equal(t, o.MemberlistConfig.AdvertiseAddr, address.External().String())
	assert.Equal(t, o.MemberlistConfig.BindPort, cfg.Cluster.Gossip)
}

func TestCluster_clusterEventLoop(t *testing.T) {
	assert.NotPanics(t, func() {
		s := new(Cluster)
		s.closing = make(chan bool)
		s.events = make(chan serf.Event, 10)
		timeout := time.After(50 * time.Millisecond)
		go func() {
			for {
				select {
				case <-timeout:
					close(s.closing)
				}
			}
		}()

		s.events <- serf.UserEvent{Name: "test event"}
		s.clusterEventLoop()
	})
}

func TestCluster_Name(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Cluster.NodeName = "hello"
	s, err := NewCluster(cfg.Cluster, nil)
	assert.NoError(t, err)

	assert.Equal(t, "hello", s.LocalName())
}
