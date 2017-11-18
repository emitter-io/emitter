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
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/http"
	"github.com/emitter-io/emitter/utils"
)

// Noop implements Storage contract.
var _ Storage = new(HTTP)

// HTTP represents a storage which uses HTTP requests to store/retrieve messages.
type HTTP struct {
	sync.Mutex
	frame message.Frame // The pending message frame.
	base  string        // The base url to use for the storage.
	http  http.Client   // The http client to use.
	done  chan bool     // The closing channel.
}

// NewHTTP creates a new HTTP storage.
func NewHTTP() *HTTP {
	return &HTTP{
		done: make(chan bool),
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
	interval := 10 * time.Millisecond
	if v, ok := config["interval"]; ok {
		if i, ok := v.(float64); ok {
			interval = time.Duration(i) * time.Millisecond
		}
	}

	// Get the url from the provider configuration
	if url, ok := config["url"]; ok {
		s.base = url.(string)
		s.http, err = http.NewClient(s.base, 30*time.Second)

		utils.Repeat(s.store, interval, s.done) // TODO: closing chan
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

	s.frame = append(s.frame, *m)
	return nil
}

// QueryLast performs a query and attempts to fetch last n messages where
// n is specified by limit argument. It returns a channel which will be
// ranged over to retrieve messages asynchronously.
func (s *HTTP) QueryLast(ssid []uint32, limit int) (ch <-chan []byte, err error) {
	re := make(chan []byte)
	ch = re // We need to return the same channel, but receive only

	// Get the raw bytes
	var resp []byte
	if resp, err = s.http.Get(s.buildLastURL(ssid, limit), nil, http.NewHeader("Accept", "application/binary")); err == nil {

		// Decode the frame we received from the server
		if frame, err := message.DecodeFrame(resp); err == nil {
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
	close(s.done)
	return nil
}

// Store periodically flushes the pending queue.
func (s *HTTP) store() {

	// Encode the frame
	s.Lock()
	buffer, _ := s.frame.Encode()
	s.frame = s.frame[:0]
	s.Unlock()

	// TODO: Make sure we don't lose messages if something happens
	if _, err := s.http.Post(s.buildAppendURL(), buffer, nil, http.NewHeader("Content-Type", "application/binary")); err != nil {
		logging.LogError("http storage", "storing messages", err)
	}
}

// Builds an append URL
func (s *HTTP) buildAppendURL() string {
	return s.base + "msg/append"
}

// Builds last query URL
func (s *HTTP) buildLastURL(ssid []uint32, limit int) string {
	enc, _ := json.Marshal(ssid)
	return s.base + fmt.Sprintf("msg/last?ssid=%s&n=%d", string(enc), limit)
}
