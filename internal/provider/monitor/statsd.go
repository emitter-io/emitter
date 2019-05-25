/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/

package monitor

import (
	"context"
	"strings"
	"time"

	"github.com/emitter-io/emitter/internal/async"
	"github.com/emitter-io/stats"
	"gopkg.in/alexcesaro/statsd.v2"
)

// Noop implements Storage contract.
var _ Storage = new(Statsd)

// Statsd represents a storage which publishes stats to a statsd sink.
type Statsd struct {
	reader stats.Snapshotter  // The reader which reads the snapshot of stats.
	client *statsd.Client     // The statsd client to use.
	cancel context.CancelFunc // The cancellation function.
	nodeID string             // The ID of the node for tagging.
}

// NewStatsd creates a new statsd sink.
func NewStatsd(snapshotter stats.Snapshotter, nodeID string) *Statsd {
	return &Statsd{
		reader: snapshotter,
		nodeID: nodeID,
	}
}

// Name returns the name of the provider.
func (s *Statsd) Name() string {
	return "statsd"
}

// Configure configures the storage. The config parameter provided is
// loosely typed, since various storage mechanisms will require different
// configurations.
func (s *Statsd) Configure(config map[string]interface{}) (err error) {

	// Get the interval from the provider configuration
	interval := defaultInterval
	if v, ok := config["interval"]; ok {
		if i, ok := v.(float64); ok {
			interval = time.Duration(i) * time.Millisecond
		}
	}

	// Get the url from the provider configuration
	url := ":8125"
	if u, ok := config["url"]; ok {
		url = u.(string)
	}

	// Create statsd client
	if s.client, err = statsd.New(
		statsd.Address(url),
		statsd.Prefix("emitter"),
		statsd.TagsFormat(statsd.Datadog),
		statsd.Tags("broker", s.nodeID),
	); err == nil {
		s.cancel = async.Repeat(context.Background(), interval, s.write)
	}
	return
}

// Flush reads and writes stats into this stats sink.
func (s *Statsd) write() {

	// Create a snapshot and restore it straight away
	snapshot := s.reader.Snapshot()
	m, err := stats.Restore(snapshot)
	if err != nil {
		return
	}

	// Send the node and process-level metrics through
	metrics := m.ToMap()
	s.gauge(metrics, "node.peers")
	s.gauge(metrics, "node.conns")
	s.gauge(metrics, "node.subs")

	// Send everything to statsd
	for name := range metrics {
		prefix := strings.Split(name, ".")[0]
		switch prefix {
		case "proc", "heap", " mcache", "mspan", "stack", "gc", "go":
			s.gauge(metrics, name)
		case "rcv", "send":
			s.histogram(metrics, name)
		}
	}

	// Flush the client as well
	s.client.Flush()
}

// Gauge sends the metric as a gauge
func (s *Statsd) gauge(source map[string]stats.Snapshot, metric string) {
	if v, ok := source[metric]; ok {
		s.client.Gauge(metric, v.Max())
	}
}

// Gauge sends the metric as a gauge
func (s *Statsd) histogram(source map[string]stats.Snapshot, metric string) {
	if v, ok := source[metric]; ok {
		for _, sample := range v.Sample {
			s.client.Histogram(metric, sample)
		}
	}
}

// Close gracefully terminates the storage and ensures that every related
// resource is properly disposed.
func (s *Statsd) Close() error {
	if s.cancel != nil {
		s.cancel()
		s.client.Close()
	}

	return nil
}
