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
	"fmt"
	"time"

	"github.com/emitter-io/address"
	"github.com/emitter-io/emitter/internal/async"
	"github.com/emitter-io/stats"
)

// Noop implements Storage contract.
var _ Storage = new(Self)

// Self represents a storage which self-publishes stats.
type Self struct {
	reader  stats.Snapshotter    // The reader which reads the snapshot of stats.
	channel string               // The channel name to publish into.
	publish func(string, []byte) // The publish function to use.
	cancel  context.CancelFunc   // The cancellation function.
}

// NewSelf creates a new self-publishing stats sink.
func NewSelf(snapshotter stats.Snapshotter, selfPublish func(string, []byte)) *Self {
	return &Self{
		publish: selfPublish,
		channel: "stats",
		reader:  snapshotter,
	}
}

// Name returns the name of the provider.
func (s *Self) Name() string {
	return "self"
}

// Configure configures the storage. The config parameter provided is
// loosely typed, since various storage mechanisms will require different
// configurations.
func (s *Self) Configure(config map[string]interface{}) error {

	// Get the interval from the provider configuration
	interval := time.Second
	if v, ok := config["interval"]; ok {
		if i, ok := v.(float64); ok {
			interval = time.Duration(i) * time.Millisecond
		}
	}

	// Get the url from the provider configuration
	if c, ok := config["channel"]; ok {
		s.channel = c.(string)
	}

	// Set channel name
	s.channel = fmt.Sprintf("%s/%s/", s.channel, address.GetHardware().Hex())

	// Setup a repeat flush
	s.cancel = async.Repeat(context.Background(), interval, s.write)
	return nil
}

// Flush reads and writes stats into this stats sink.
func (s *Self) write() {
	if snapshot := s.reader.Snapshot(); len(snapshot) > 0 {
		s.publish(s.channel, snapshot)
	}
}

// Close gracefully terminates the storage and ensures that every related
// resource is properly disposed.
func (s *Self) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}
