package monitor

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
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
	})
	assert.NoError(t, err)
	assert.NotPanics(t, func() {
		s.write()
	})
}

func TestPrometheus_Request(t *testing.T) {

	m := stats.New()
	for i := int32(0); i < 100; i++ {
		m.Measure("proc.test", i)
		m.Measure("node.test", i)
		m.Measure("rcv.test", i/10)
		m.Measure("node.peers", 2)
		m.Measure("node.conns", i)
		m.Measure("node.subs", i)
	}

	mux := http.NewServeMux()
	s := NewPrometheus(m, mux)
	defer s.Close()

	err := s.Configure(map[string]interface{}{
		"interval": 1000000.00,
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	// assert gauges
	assert.Contains(t, string(content), "node_peers 2")
	assert.Contains(t, string(content), "node_subs 99")
	assert.Contains(t, string(content), "node_conns 99")

	// assert histograms
	assert.Contains(t, string(content), "rcv_test_bucket{le=\"0.01\"} 10")
	assert.Contains(t, string(content), "rcv_test_sum 450")
	assert.Contains(t, string(content), "rcv_test_count 100")

	// from InstrumentMetricHandler
	assert.Contains(t, string(content), "promhttp_metric_handler_requests_total")

	// from the NewGoCollector
	assert.Contains(t, string(content), "go_threads")
}
