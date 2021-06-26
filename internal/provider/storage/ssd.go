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
	"context"
	"os"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/emitter-io/emitter/internal/async"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/service"
	"github.com/kelindar/binary"
)

// ------------------------------------------------------------------------------------ //

// SSD represents an SSD-optimized storage storage.
type SSD struct {
	retain uint32             // The configured TTL for 'retained' messages.
	survey service.Surveyor   // The cluster surveyor.
	db     *badger.DB         // The underlying database to use for messages.
	cancel context.CancelFunc // The cancellation function.
}

// NewSSD creates a new SSD-optimized storage storage.
func NewSSD(survey service.Surveyor) *SSD {
	return &SSD{
		survey: survey,
	}
}

// Name returns the name of the provider.
func (s *SSD) Name() string {
	return "ssd"
}

// Configure configures the storage. The config parameter provided is
// loosely typed, since various storage mechanisms will require different
// configurations.
func (s *SSD) Configure(config map[string]interface{}) error {

	// Get the interval from the provider configuration
	dir := "/data"
	if v, ok := config["dir"]; ok {
		if d, ok := v.(string); ok {
			dir = d
		}
	}

	// Make sure we have a directory
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

	// Create the options
	opts := badger.DefaultOptions(dir)
	opts.SyncWrites = false

	//opts.ValueLogLoadingMode = options.FileIO

	// Attempt to open the database
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}

	// Setup the database and start GC
	s.db = db
	s.retain = configUint32(config, "retain", defaultRetain)
	s.cancel = async.Repeat(context.Background(), 30*time.Minute, s.GC)
	return nil
}

// Store appends the messages to the store.
func (s *SSD) Store(m *message.Message) error {
	if m.TTL == message.RetainedTTL {
		m.TTL = s.retain
	}

	// TODO: add batching instead of storing one by one
	return s.storeFrame(message.Frame{*m})
}

// storeFrame appends the frame of messages to the store.
func (s *SSD) storeFrame(msgs message.Frame) error {
	encoded := encodeFrame(msgs)
	return s.db.Update(func(tx *badger.Txn) error {
		for _, m := range encoded {
			entry := m // Copy address
			tx.SetEntry(entry)
		}
		return nil
	})
}

// encodeMessage encodes a message frame so we can store it
func encodeFrame(msgs message.Frame) []*badger.Entry {
	entries := make([]*badger.Entry, 0, len(msgs))
	for _, m := range msgs {
		entries = append(entries, &badger.Entry{
			Key:       m.ID,
			Value:     m.Encode(),
			ExpiresAt: uint64(m.Expires().Unix()),
		})
	}
	return entries
}

// Query performs a query and attempts to fetch last n messages where
// n is specified by limit argument. From and until times can also be specified
// for time-series retrieval.
func (s *SSD) Query(ssid message.Ssid, from, until time.Time, limit int) (message.Frame, error) {

	// Construct a query and lookup locally first
	query := newLookupQuery(ssid, from, until, limit)
	match := s.lookup(query)

	// Issue the message survey to the cluster
	if req, err := binary.Marshal(query); err == nil && s.survey != nil {
		if awaiter, err := s.survey.Query("ssdstore", req); err == nil {

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
func (s *SSD) OnSurvey(surveyType string, payload []byte) ([]byte, bool) {
	if surveyType != "ssdstore" {
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

	//logging.LogTarget("ssd", surveyType+" survey received", query)

	// Send back the response
	f := s.lookup(query)
	b := f.Encode()
	return b, true
}

// Lookup performs a against the storage.
func (s *SSD) lookup(q lookupQuery) (matches message.Frame) {
	matches = make(message.Frame, 0, q.Limit)
	if err := s.db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(badger.IteratorOptions{
			PrefetchValues: false,
		})
		defer it.Close()

		// Since we're starting backwards, seek to the 'until' position first and then
		// we'll iterate forward but have reverse time ('until' -> 'from')
		prefix := message.NewPrefix(q.Ssid, q.Until)

		// Seek the prefix and check the key so we can quickly exit the iteration.
		for it.Seek(prefix); it.Valid() &&
			message.ID(it.Item().Key()).HasPrefix(q.Ssid, q.From) &&
			len(matches) < q.Limit; it.Next() {
			if message.ID(it.Item().Key()).Match(q.Ssid, q.From, q.Until) {
				if msg, err := loadMessage(it.Item()); err == nil {
					matches = append(matches, msg)
				}
			}
		}

		return nil
	}); err != nil {
		logging.LogError("ssd", "query lookup", err)
	}
	return
}

// Close is used to gracefully close the connection.
func (s *SSD) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	return s.db.Close()
}

// LoadMessage loads the message from badger item.
func loadMessage(item *badger.Item) (message.Message, error) {
	data, err := item.ValueCopy(nil)
	if err != nil {
		return message.Message{}, err
	}

	return message.DecodeMessage(data)
}

// GC runs the garbage collection on the storage
func (s *SSD) GC() {
	s.db.RunValueLogGC(0.50)
}
