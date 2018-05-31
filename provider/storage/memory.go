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

package storage

import (
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/message"
	"github.com/karlseguin/ccache"
	"github.com/kelindar/binary"
)

// InMemory implements Storage contract.
var _ Storage = new(InMemory)

// InMemory represents a storage which does nothing.
type InMemory struct {
	cluster Surveyor      // The cluster surveyor.
	cur     *sync.Map     // The cursor map which stores the last written offset.
	mem     *ccache.Cache // The LRU cache with TTL.
}

// NewInMemory creates a new in-memory storage.
func NewInMemory(cluster Surveyor) *InMemory {
	return &InMemory{
		cluster: cluster,
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
	key := m.Ssid().Encode()
	trunk := key[:16]

	// Get and increment the last message cursor
	cur, _ := s.cur.LoadOrStore(trunk, new(uint64))
	idx := atomic.AddUint64(cur.(*uint64), 1)

	// Set the key in form of (ssid:index) so we can retrieve
	s.mem.Set(fmt.Sprintf("%v:%v", trunk, idx), *m, time.Duration(m.TTL)*time.Second)

	//logging.LogTarget("memstore", "message stored", idx)
	return nil
}

// Query performs a query and attempts to fetch last n messages where
// n is specified by limit argument. From and until times can also be specified
// for time-series retrieval.
func (s *InMemory) Query(ssid message.Ssid, from, until time.Time, limit int) (message.Frame, error) {

	// Construct a query and lookup locally first
	query := newLookupQuery(ssid, from, until, limit)
	match := s.lookup(query)

	// Issue the message survey to the cluster
	if req, err := binary.Marshal(query); err == nil && s.cluster != nil {
		if awaiter, err := s.cluster.Survey("memstore", req); err == nil {

			// Wait for all presence updates to come back (or a deadline)
			for _, resp := range awaiter.Gather(2000 * time.Millisecond) {
				if frame, err := message.DecodeFrame(resp); err == nil {
					match = append(match, frame...)
				}
			}
		}
	}

	match.Sort()
	match.Limit(limit)
	return match, nil
}

// OnSurvey handles an incoming cluster lookup request.
func (s *InMemory) OnSurvey(surveyType string, payload []byte) ([]byte, bool) {
	if surveyType != "memstore" {
		return nil, false
	}

	// Decode the request
	var query lookupQuery
	if err := binary.Unmarshal(payload, &query); err != nil {
		return nil, false
	}

	// Check if the SSID is properly constructed
	if len(query.Ssid) < 2 {
		return nil, false
	}

	//logging.LogTarget("memstore", surveyType+" survey received", query)

	// Send back the response
	f := s.lookup(query)
	b := f.Encode()
	return b, true
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
				if query.MatchString(msg.Ssid().Encode()) {
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
