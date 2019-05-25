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

package usage

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/emitter-io/config"
	"github.com/emitter-io/emitter/internal/async"
	"github.com/emitter-io/emitter/internal/network/http"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/kelindar/binary"
)

// Metering represents a contract for a usage metering
type Metering interface {
	config.Provider

	// Get retrieves a meter for a contract.
	Get(id uint32) Meter
}

// ------------------------------------------------------------------------------------

// Noop implements Metering contract.
var _ Metering = new(NoopStorage)

// NoopStorage represents a usage storage which does nothing.
type NoopStorage struct{}

// NewNoop creates a new no-op storage.
func NewNoop() *NoopStorage {
	return new(NoopStorage)
}

// Name returns the name of the provider.
func (s *NoopStorage) Name() string {
	return "noop"
}

// Configure configures the provider
func (s *NoopStorage) Configure(config map[string]interface{}) error {
	return nil
}

// Get retrieves a meter for a contract.
func (s *NoopStorage) Get(id uint32) Meter {
	return NewMeter(id)
}

// ------------------------------------------------------------------------------------

// HTTPStorage implements Metering contract.
var _ Metering = new(HTTPStorage)

// HTTPStorage represents a usage storage which posts meters over HTTP.
type HTTPStorage struct {
	counters *sync.Map          // The counters map.
	url      string             // The url to post to.
	http     http.Client        // The http client to use.
	head     []http.HeaderValue // The http headers to add with each request.
	cancel   context.CancelFunc // The cancellation function.
}

// NewHTTP creates a new HTTP storage
func NewHTTP() *HTTPStorage {
	return &HTTPStorage{
		counters: new(sync.Map),
	}
}

// Name returns the name of the provider.
func (s *HTTPStorage) Name() string {
	return "http"
}

// Configure configures the provider.
func (s *HTTPStorage) Configure(config map[string]interface{}) (err error) {
	if config == nil {
		return errors.New("Configuration was not provided for HTTP metering provider")
	}

	// Get the interval from the provider configuration
	interval := time.Second
	if v, ok := config["interval"]; ok {
		if i, ok := v.(float64); ok {
			interval = time.Duration(i) * time.Millisecond
		}
	}

	// Set accept header for the metering
	headers := []http.HeaderValue{http.NewHeader("Accept", "application/binary")}

	// Get the authorization header to add to the request
	if v, ok := config["authorization"]; ok {
		if header, ok := v.(string); ok {
			headers = append(headers, http.NewHeader("Authorization", header))
		}
	}

	// Get the url from the provider configuration
	if url, ok := config["url"]; ok {
		s.url = url.(string)
		s.http, err = http.NewClient(30 * time.Second)
		s.cancel = async.Repeat(context.Background(), interval, s.store)
		return
	}

	return errors.New("The 'url' parameter was not provider in the configuration for HTTP contract provider")
}

// Get retrieves a meter for a contract.
func (s *HTTPStorage) Get(id uint32) Meter {
	meter, _ := s.counters.LoadOrStore(id, NewMeter(id))
	return meter.(Meter)
}

// Close closes the storage.
func (s *HTTPStorage) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

// Store periodically stores the counters by sending them through HTTP.
func (s *HTTPStorage) store() {
	counters := make([]encodedUsage, 0)
	s.counters.Range(func(k, v interface{}) bool {
		counters = append(counters, v.(*usage).reset())
		return true
	})

	// Encode as binary and post without waiting for the body
	if encoded, err := binary.Marshal(counters); err == nil {
		if _, err := s.http.Post(s.url, encoded, nil, s.head...); err != nil {
			logging.LogError("http metering", "reporting counters", err)
		}
	}
}
