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
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/utils"
	"github.com/karlseguin/ccache"
)

var (
	errNotFound = errors.New("No messages were found")
)

// The lookup query to send out to the cluster.
type lookupQuery struct {
	Ssid  []uint32 // The ssid to match.
	Limit int      // The maximum number of elements to return.
}

// InMemory implements Storage contract.
var _ Storage = new(InMemory)

// InMemory represents a storage which does nothing.
type InMemory struct {
	cur   *sync.Map                                     // The cursor map which stores the last written offset.
	mem   *ccache.Cache                                 // The LRU cache with TTL.
	Query func(string, []byte) (message.Awaiter, error) // The cluster request function.
}

// NewInMemory creates a new in-memory storage.
func NewInMemory(q func(string, []byte) (message.Awaiter, error)) *InMemory {
	return &InMemory{
		Query: q,
	}
}

// Name returns the name of the provider.
func (s *InMemory) Name() string {
	return "inmemory"
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
func (s *InMemory) Store(m *message.Message) error {

	// Get the string version of the SSID trunk
	key := message.Ssid(m.Ssid).Encode()
	trunk := key[:16]

	// Get and increment the last message cursor
	cur, _ := s.cur.LoadOrStore(trunk, new(uint64))
	idx := atomic.AddUint64(cur.(*uint64), 1)

	// If no time was set, add it
	if m.Time == 0 {
		m.Time = time.Now().UnixNano()
	}

	// Set the key in form of (ssid:index) so we can retrieve
	s.mem.Set(fmt.Sprintf("%v:%v", trunk, idx), *m, time.Duration(m.TTL)*time.Second)

	//logging.LogTarget("memstore", "message stored", idx)
	return nil
}

// QueryLast performs a query and attempts to fetch last n messages where
// n is specified by limit argument. It returns a channel which will be
// ranged over to retrieve messages asynchronously.
func (s *InMemory) QueryLast(ssid []uint32, limit int) (<-chan []byte, error) {

	// Construct a query and lookup locally first
	query := lookupQuery{Ssid: ssid, Limit: limit}
	match := s.lookup(query)

	// Issue the presence query to the cluster
	if req, err := utils.Encode(query); err == nil && s.Query != nil {
		if awaiter, err := s.Query("memstore", req); err == nil {

			// Wait for all presence updates to come back (or a deadline)
			for _, resp := range awaiter.Gather(2000 * time.Millisecond) {
				if frame, err := message.DecodeFrame(resp); err == nil {
					match = append(match, frame...)
				}
			}
		}
	}

	// Sort the matches by time
	sort.Slice(match, func(i, j int) bool { return match[i].Time < match[j].Time })

	// Set the offset
	i := len(match) - limit
	if i < 0 {
		i = 0
	}

	// Project to return only payloads
	ch := make(chan []byte, limit)
	match = match[i:]
	for _, msg := range match {
		ch <- msg.Payload
	}

	// Close and return the channel
	close(ch)
	return ch, nil
}

// OnRequest handles an incoming cluster lookup request.
func (s *InMemory) OnRequest(queryType string, payload []byte) ([]byte, bool) {
	if queryType != "memstore" {
		return nil, false
	}

	// Decode the request
	var query lookupQuery
	if err := utils.Decode(payload, &query); err != nil {
		return nil, false
	}

	// Check if the SSID is properly constructed
	if len(query.Ssid) < 2 {
		return nil, false
	}

	//logging.LogTarget("memstore", queryType+" query received", query)

	// Send back the response
	f := s.lookup(query)
	b, err := f.Encode()
	return b, err == nil
}

// Lookup performs a against the cache.
func (s *InMemory) lookup(q lookupQuery) (matches message.Frame) {
	matches = make(message.Frame, 0, q.Limit)
	matchCount := 0

	// Get the string version of the SSID trunk
	key := message.Ssid(q.Ssid).Encode()
	trunk := key[:16]

	// Get the value of the last message cursor
	last, ok := s.cur.Load(trunk)
	if !ok {
		return
	}

	// Create a compiled regular expression for querying
	if query, err := regexp.Compile(key + ".*"); err == nil {

		// Iterate from last to 0 (limit to last n) and append locally
		for i := atomic.LoadUint64(last.(*uint64)); i > 0 && matchCount < q.Limit; i-- {
			if item := s.mem.Get(fmt.Sprintf("%v:%v", trunk, i)); item != nil && !item.Expired() {
				msg := item.Value().(message.Message)

				// Match using regular expression
				if query.MatchString(msg.Ssid.Encode()) {
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
