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
			subscriptions: message.NewTrie(),
			measurer:      stats.NewNoop(),
			License:       license,
		}
		defer s.Close()

		var out []byte
		sampler := newSampler(s, s.measurer)
		n, err := sampler.Read(out)
		assert.Equal(t, 0, n)
		assert.Equal(t, "EOF", err.Error())
	})
}
