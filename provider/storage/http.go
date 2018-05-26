/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
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

package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/emitter-io/emitter/async"
	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/network/http"
	"github.com/emitter-io/emitter/provider/logging"
)

// Noop implements Storage contract.
var _ Storage = new(HTTP)

// Default message frame size to use
const defaultFrameSize = 128

// HTTP represents a storage which uses HTTP requests to store/retrieve messages.
type HTTP struct {
	sync.Mutex
	frame  message.Frame      // The pending message frame.
	base   string             // The base url to use for the storage.
	http   http.Client        // The http client to use.
	head   []http.HeaderValue // The http headers to add with each request.
	cancel context.CancelFunc // The cancellation function.
}

// NewHTTP creates a new HTTP storage.
func NewHTTP() *HTTP {
	return &HTTP{
		frame: make(message.Frame, 0, 64),
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
	interval := 25 * time.Millisecond
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
		s.base = url.(string)
		s.http, err = http.NewClient(s.base, 30*time.Second)
		s.head = headers

		s.cancel = async.Repeat(context.Background(), interval, s.store)
		return
	}

	return errors.New("The 'url' parameter was not provider in the configuration for HTTP storage")
}

// Store is used to store a message, the SSID provided must be a full SSID
// SSID, where first element should be a contract ID. The time resolution
// for TTL will be in seconds. The function is executed synchronously and
// it returns an error if some error was encountered during storage.
func (s *HTTP) Store(m *message.Message) error {
	s.Lock()
	defer s.Unlock()

	// If no time was set, add it
	if m.Time == 0 {
		m.Time = time.Now().UnixNano()
	}

	// Append to the frame
	s.frame = append(s.frame, *m)
	return nil
}

// storeMany stores an entire message frame without changing it
func (s *HTTP) storeMany(frame message.Frame) {
	s.Lock()
	defer s.Unlock()

	s.frame = append(s.frame, frame...)
}

// QueryLast performs a query and attempts to fetch last n messages where
// n is specified by limit argument. It returns a channel which will be
// ranged over to retrieve messages asynchronously.
func (s *HTTP) QueryLast(ssid []uint32, limit int) (ch <-chan []byte, err error) {
	re := make(chan []byte, limit)
	ch = re // We need to return the same channel, but receive only

	// Get the raw bytes
	var resp []byte
	if resp, err = s.http.Get(s.buildLastURL(ssid, limit), nil, s.head...); err == nil {

		// Decode the frame we received from the server
		var frame message.Frame
		if frame, err = message.DecodeFrame(resp); err == nil {
			for _, msg := range frame {
				re <- msg.Payload
			}
			close(re)
		}
	}

	// If there's an error, return a closed channel
	if err != nil {
		close(re)
	}

	return
}

// Close gracefully terminates the storage and ensures that every related
// resource is properly disposed.
func (s *HTTP) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

// Store periodically flushes the pending queue.
func (s *HTTP) store() {
	if len(s.frame) == 0 {
		return
	}

	// Swap the frame and encode it
	frame := s.swap()
	buffer := frame.Encode()

	// Write messages through HTTP Post and append them back on error
	if _, err := s.http.Post(s.buildAppendURL(), buffer, nil, s.head...); err != nil {
		logging.LogError("http storage", "storing messages", err)
		s.storeMany(frame)
	}
}

// swap swaps the frame and returns the frame we can encode.
func (s *HTTP) swap() (swapped message.Frame) {
	s.Lock()
	defer s.Unlock()

	swapped = s.frame
	s.frame = message.NewFrame(defaultFrameSize)
	return
}

// Builds an append URL
func (s *HTTP) buildAppendURL() string {
	return s.base + "v1/add/"
}

// Builds last query URL
func (s *HTTP) buildLastURL(ssid []uint32, limit int) string {
	enc, _ := json.Marshal(ssid)
	return s.base + fmt.Sprintf("v1/get/?ssid=%s&limit=%d", string(enc), limit)
}
