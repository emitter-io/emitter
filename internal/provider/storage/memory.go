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

package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/service"
	"github.com/kelindar/binary"
	"github.com/tidwall/buntdb"
)

// InMemory implements Storage contract.
var _ Storage = new(InMemory)

// InMemory represents a storage which does nothing.
type InMemory struct {
	retain uint32           // The configured TTL for 'retained' messages.
	survey service.Surveyor // The surveyor to use.
	index  *sync.Map        // The set of indices.
	db     *buntdb.DB       // The in-memory storage.
}

// NewInMemory creates a new in-memory storage.
func NewInMemory(survey service.Surveyor) *InMemory {
	return &InMemory{
		survey: survey,
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
	s.index = new(sync.Map)
	db, err := buntdb.Open(":memory:")
	if err == nil {
		s.db = db
	}

	s.retain = configUint32(config, "retain", defaultRetain)
	return err
}

// Store is used to store a message, the SSID provided must be a full SSID
// SSID, where first element should be a contract ID. The time resolution
// for TTL will be in seconds. The function is executed synchronously and
// it returns an error if some error was encountered during storage.
func (s *InMemory) Store(m *message.Message) error {
	if m.TTL == message.RetainedTTL {
		m.TTL = s.retain
	}

	// Get the string version of the SSID trunk
	msg := string(m.Encode())
	idx := fmt.Sprintf("%x", m.ID[:4])

	// Make sure we have an index
	if _, loaded := s.index.LoadOrStore(idx, true); !loaded {
		s.db.Update(func(tx *buntdb.Tx) error {
			return tx.CreateIndex(idx, fmt.Sprintf("%s:*", idx), indexMessage)
		})
	}

	// Write the message
	return s.db.Update(func(tx *buntdb.Tx) error {
		tx.Set(fmt.Sprintf("%s:%s", idx, m.ID), msg, &buntdb.SetOptions{
			Expires: m.TTL > 0,
			TTL:     time.Second * time.Duration(m.TTL),
		})
		return nil
	})
}

// Query performs a query and attempts to fetch last n messages where
// n is specified by limit argument. From and until times can also be specified
// for time-series retrieval.
func (s *InMemory) Query(ssid message.Ssid, from, until time.Time, limit int) (message.Frame, error) {

	// Construct a query and lookup locally first
	query := newLookupQuery(ssid, from, until, limit)
	match := s.lookup(query)

	// Issue the message survey to the cluster
	if req, err := binary.Marshal(query); err == nil && s.survey != nil {
		if awaiter, err := s.survey.Query("memstore", req); err == nil {

			// Wait for all presence updates to come back (or a deadline)
			for _, resp := range awaiter.Gather(2000 * time.Millisecond) {
				if frame, err := message.DecodeFrame(resp); err == nil {
					match = append(match, frame...)
				}
			}
		}
	}

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
	prefix := message.NewPrefix(q.Ssid, q.From)
	idx := fmt.Sprintf("%x", prefix[:4])

	matches = make(message.Frame, 0, q.Limit)
	matchCount := 0
	s.db.View(func(tx *buntdb.Tx) error {
		tx.Ascend(idx, func(key, value string) bool {
			id := message.ID(key[9:])
			if id.Match(q.Ssid, q.From, q.Until) {
				if msg, err := message.DecodeMessage([]byte(value)); err == nil {

					matchCount++
					matches = append(matches, msg)
					if matchCount >= q.Limit {
						return false
					}
				}
			}

			return true
		})
		return nil
	})

	// Return the matching messages we found
	return
}

// Close gracefully terminates the storage and ensures that every related
// resource is properly disposed.
func (s *InMemory) Close() error {
	return nil
}

// indexMessage sorts two buntdb messages.
func indexMessage(a, b string) bool {
	return true // Do not sort by value
}
