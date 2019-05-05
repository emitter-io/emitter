package monitor

import (
	"net/http"
	"testing"

	"github.com/emitter-io/stats"
	"github.com/stretchr/testify/assert"
)

func TestPrometheus_HappyPath(t *testing.T) {
	m := stats.New()
	for i := int32(0); i < 100; i++ {
		m.Measure("proc.test", i)
		m.Measure("node.test", i)
		m.Measure("rcv.test", i)
	}

	mux := http.NewServeMux()

	s := NewPrometheus(m, mux)
	defer s.Close()

	err := s.Configure(map[string]interface{}{
		"interval": 1000000.00,
		"url":      ":8125",
	})
	assert.NoError(t, err)
	assert.NotPanics(t, func() {
		s.write()
	})
}
