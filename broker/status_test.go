package broker

import (
	"io/ioutil"
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
			measurer:      stats.New(),
			License:       license,
		}
		defer s.Close()

		sampler := newSampler(s, s.measurer)

		// Read everything
		b, err := ioutil.ReadAll(sampler)
		assert.NotZero(t, len(b))
		assert.NoError(t, err)

	})
}
