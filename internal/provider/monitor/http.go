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
	"errors"
	"time"

	"github.com/emitter-io/emitter/internal/async"
	"github.com/emitter-io/emitter/internal/network/http"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/stats"
)

// Noop implements Storage contract.
var _ Storage = new(Self)

// HTTP represents a storage which publishes stats over HTTP.
type HTTP struct {
	reader stats.Snapshotter  // The reader which reads the snapshot of stats.
	url    string             // The base url to use for the storage.
	http   http.Client        // The http client to use.
	head   []http.HeaderValue // The http headers to add with each request.
	cancel context.CancelFunc // The cancellation function.
}

// NewHTTP creates a new HTTP stats sink.
func NewHTTP(snapshotter stats.Snapshotter) *HTTP {
	return &HTTP{
		reader: snapshotter,
	}
}

// Name returns the name of the provider.
func (s *HTTP) Name() string {
	return "http"
}

// Configure configures the storage. The config parameter provided is
// loosely typed, since various storage mechanisms will require different
// configurations.
func (s *HTTP) Configure(config map[string]interface{}) (err error) {
	if config == nil {
		return errors.New("Configuration was not provided for HTTP storage")
	}

	// Get the interval from the provider configuration
	interval := defaultInterval
	if v, ok := config["interval"]; ok {
		if i, ok := v.(float64); ok {
			interval = time.Duration(i) * time.Millisecond
		}
	}

	// Get the authorization header to add to the request
	headers := []http.HeaderValue{http.NewHeader("Accept", "application/binary")}
	if v, ok := config["authorization"]; ok {
		if header, ok := v.(string); ok {
			headers = append(headers, http.NewHeader("Authorization", header))
		}
	}

	// Get the url from the provider configuration
	if url, ok := config["url"]; ok {
		s.url = url.(string)
		s.http, err = http.NewClient(30 * time.Second)
		s.head = headers
		s.cancel = async.Repeat(context.Background(), interval, s.write)
		return
	}

	return errors.New("The 'url' parameter was not provider in the configuration for HTTP storage")
}

// Flush reads and writes stats into this stats sink.
func (s *HTTP) write() {
	if snapshot := s.reader.Snapshot(); len(snapshot) > 0 {
		if _, err := s.http.Post(s.url, snapshot, nil, s.head...); err != nil {
			logging.LogError("http stats", "sending stats", err)
		}
	}
}

// Close gracefully terminates the storage and ensures that every related
// resource is properly disposed.
func (s *HTTP) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}
