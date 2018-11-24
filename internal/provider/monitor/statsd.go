/**********************************************************************************
* Copyright (c) 2009-2018 Misakai Ltd.
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
}

// NewStatsd creates a new statsd sink.
func NewStatsd(snapshotter stats.Snapshotter) *Statsd {
	return &Statsd{
		reader: snapshotter,
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
	if s.client, err = statsd.New(statsd.Address(url), statsd.Prefix("emitter")); err == nil {
		s.cancel = async.Repeat(context.Background(), interval, s.write)
	}

	return
}

// Flush reads and writes stats into this stats sink.
func (s *Statsd) write() {

	// Create a snapshot and restore it straight away
	snapshot := s.reader.Snapshot()
	metrics, err := stats.Restore(snapshot)
	if err != nil {
		return
	}

	// Send everything to statsd
	for _, v := range metrics {
		q := v.Quantile(25, 50, 75, 90, 95, 99)
		s.client.Gauge(v.Name()+".p25", q[0])
		s.client.Gauge(v.Name()+".p50", q[1])
		s.client.Gauge(v.Name()+".p75", q[2])
		s.client.Gauge(v.Name()+".p90", q[3])
		s.client.Gauge(v.Name()+".p95", q[4])
		s.client.Gauge(v.Name()+".p99", q[5])
		s.client.Gauge(v.Name()+".min", v.Min())
		s.client.Gauge(v.Name()+".max", v.Max())
		s.client.Gauge(v.Name()+".avg", v.Mean())
		s.client.Gauge(v.Name()+".var", v.Variance())
		s.client.Gauge(v.Name()+".stddev", v.StdDev())
		s.client.Count(v.Name()+".count", v.Count())
	}

	// Flush the client as well
	s.client.Flush()
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
