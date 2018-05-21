package broker

import (
	"testing"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/security"
	"github.com/emitter-io/stats"
	"github.com/stretchr/testify/assert"
)

func Test_sendStats(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)

	assert.NotPanics(t, func() {
		s := &Service{
			Closing:       make(chan bool),
			measurer:      stats.NewNoop(),
			subscriptions: message.NewTrie(),
			License:       license,
		}
		defer s.Close()
		s.sendStats()
	})
}
