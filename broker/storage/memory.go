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
	"errors"
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/broker/subscription"
	"github.com/karlseguin/ccache"
)

var (
	errNotFound = errors.New("No messages were found")
)

// Message represents a stored message.
type message struct {
	ssid    string // The hex-encoded SSID
	payload []byte // The payload
}

// Size returns the byte size of the message.
func (m message) Size() int64 {
	return int64(len(m.payload))
}

// InMemory implements Storage contract.
var _ Storage = new(InMemory)

// InMemory represents a storage which does nothing.
type InMemory struct {
	cur *sync.Map
	mem *ccache.Cache
}

// Configure configures the storage. The config parameter provided is
// loosely typed, since various storage mechanisms will require different
// configurations.
func (s *InMemory) Configure(config map[string]interface{}) error {
	cfg := ccache.Configure().
		MaxSize(param(config, "maxsize", 1*1024*1024*1024)).
		ItemsToPrune(uint32(param(config, "prune", 100)))
	s.mem = ccache.New(cfg)
	s.cur = new(sync.Map)
	return nil
}

// Store is used to store a message, the SSID provided must be a full SSID
// SSID, where first element should be a contract ID. The time resolution
// for TTL will be in seconds. The function is executed synchronously and
// it returns an error if some error was encountered during storage.
func (s *InMemory) Store(ssid []uint32, payload []byte, ttl time.Duration) error {

	// Get the string version of the SSID trunk
	key := subscription.Ssid(ssid).Encode()
	trunk := key[:16]

	// Get and increment the last message cursor
	cur, _ := s.cur.LoadOrStore(trunk, new(uint64))
	idx := atomic.AddUint64(cur.(*uint64), 1)

	// Set the key in form of (ssid:index) so we can retrieve
	s.mem.Set(fmt.Sprintf("%v:%v", trunk, idx), message{ssid: key, payload: payload}, ttl)
	return nil
}

// QueryLast performs a query and attempts to fetch last n messages where
// n is specified by limit argument. It returns a channel which will be
// ranged over to retrieve messages asynchronously.
func (s *InMemory) QueryLast(ssid []uint32, limit int) (<-chan []byte, error) {
	ch := make(chan []byte)

	// 3. Query all other services and await their response
	// 4. Merge everything
	// 5. Send the entire thing through the channel

	close(ch) // Close the channel so we can return a closed one.
	return ch, nil
}

// Lookup performs a query agains the cache.
func (s *InMemory) lookup(ssid []uint32, limit int) (matches []message) {
	matches = make([]message, 0, limit)
	matchCount := 0

	// Get the string version of the SSID trunk
	key := subscription.Ssid(ssid).Encode()
	trunk := key[:16]

	// Get the value of the last message cursor
	last, ok := s.cur.Load(trunk)
	if !ok {
		return
	}

	// Create a compiled regular expression for querying
	if query, err := regexp.Compile(key + ".*"); err == nil {

		// Iterate from last to 0 (limit to last n) and append locally
		for i := atomic.LoadUint64(last.(*uint64)); i > 0 && matchCount < limit; i-- {
			if item := s.mem.Get(fmt.Sprintf("%v:%v", trunk, i)); item != nil && !item.Expired() {
				msg := item.Value().(message)

				// Match using regular expression
				if query.MatchString(msg.ssid) {
					matchCount++
					matches = append(matches, msg)
				}
			}
		}
	}

	// Return the matching messages we found
	return
}

// Close gracefully terminates the storage and ensures that every related
// resource is properly disposed.
func (s *InMemory) Close() error {
	return nil
}

// Param retrieves a parameter from the configuration.
func param(config map[string]interface{}, name string, defaultValue int64) int64 {
	if v, ok := config[name]; ok {
		if i, ok := v.(float64); ok {
			return int64(i)
		}
	}
	return defaultValue
}
