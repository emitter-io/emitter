package broker

import (
	"testing"

	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/stats"
	"github.com/stretchr/testify/assert"
)

func Test_sendStats(t *testing.T) {
	license, _ := security.ParseLicense(testLicense)

	assert.NotPanics(t, func() {
		s := &Service{
			subscriptions: message.NewTrie(),
			measurer:      stats.New(),
			License:       license,
			Config: &config.Config{
				ListenAddr: ":1234",
			},
		}
		defer s.Close()

		sampler := newSampler(s, s.measurer)
		b := sampler.Snapshot()
		assert.NotZero(t, len(b))

	})
}
