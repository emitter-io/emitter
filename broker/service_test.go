package broker

import (
	"testing"
	"time"

	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/network/address"
	"github.com/hashicorp/serf/serf"
	"github.com/stretchr/testify/assert"
)

func TestService_clusterConfig(t *testing.T) {
	cfg := config.NewDefault()
	s := new(Service)

	o := s.clusterConfig(cfg)
	assert.Equal(t, o.MemberlistConfig.SecretKey, cfg.Cluster.Key())
	assert.Equal(t, o.MemberlistConfig.AdvertiseAddr, address.External().String())
	assert.Equal(t, o.MemberlistConfig.BindPort, cfg.Cluster.Gossip)
}

func TestService_clusterEventLoop(t *testing.T) {
	assert.NotPanics(t, func() {
		s := new(Service)
		s.Closing = make(chan bool)
		s.events = make(chan serf.Event, 10)
		timeout := time.After(50 * time.Millisecond)
		go func() {
			for {
				select {
				case <-timeout:
					s.Close()
				}
			}
		}()

		s.events <- serf.UserEvent{Name: "test event"}
		s.clusterEventLoop()
	})
}

func TestService_Name(t *testing.T) {
	s := new(Service)
	assert.Equal(t, "", s.Name())

	cfg := config.NewDefault()
	cfg.Cluster.NodeName = "hello"
	_ = s.clusterConfig(cfg)
	assert.Equal(t, "hello", s.Name())
}
