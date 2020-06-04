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
	"errors"
	"io"
	"time"

	"github.com/emitter-io/config"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
)

var (
	errNotFound = errors.New("no messages were found")
)

const (
	defaultRetain = 2592000 // 30-days
)

// Storage represents a message storage contract that message storage provides
// must fulfill.
type Storage interface {
	config.Provider
	io.Closer

	// Store is used to store a message, the SSID provided must be a full SSID
	// SSID, where first element should be a contract ID. The time resolution
	// for TTL will be in seconds. The function is executed synchronously and
	// it returns an error if some error was encountered during storage.
	Store(m *message.Message) error

	// Query performs a query and attempts to fetch last n messages where
	// n is specified by limit argument. From and until times can also be specified
	// for time-series retrieval.
	Query(ssid message.Ssid, from, until time.Time, limit int) (message.Frame, error)
}

// ------------------------------------------------------------------------------------

// window constructs a time window
func window(from, until time.Time) (int64, int64) {
	t0 := from.Unix()
	t1 := until.Unix()
	if t1 == 0 {
		t1 = int64(security.MaxTime)
	}

	return t0, t1
}

// The lookup query to send out to the cluster.
type lookupQuery struct {
	Ssid  message.Ssid // The ssid to match.
	From  int64        // The beginning of the time window.
	Until int64        // The end of the time window.
	Limit int          // The maximum number of elements to return.
}

// newLookupQuery creates a new lookup query
func newLookupQuery(ssid message.Ssid, from, until time.Time, limit int) lookupQuery {
	t0, t1 := window(from, until)
	return lookupQuery{
		Ssid:  ssid,
		From:  t0,
		Until: t1,
		Limit: limit,
	}
}

// configUint32 retrieves an uint32 from the config
func configUint32(config map[string]interface{}, name string, defaultValue uint32) uint32 {
	if v, ok := config[name]; ok {
		if i, ok := v.(float64); ok && i > 0 {
			return uint32(i)
		}
	}
	return defaultValue
}

// ------------------------------------------------------------------------------------

// Noop implements Storage contract.
var _ Storage = new(Noop)

// Noop represents a storage which does nothing.
type Noop struct{}

// NewNoop creates a new no-op storage.
func NewNoop() *Noop {
	return new(Noop)
}

// Name returns the name of the provider.
func (s *Noop) Name() string {
	return "noop"
}

// Configure configures the storage. The config parameter provided is
// loosely typed, since various storage mechanisms will require different
// configurations.
func (s *Noop) Configure(config map[string]interface{}) error {
	return nil
}

// Store is used to store a message, the SSID provided must be a full SSID
// SSID, where first element should be a contract ID. The time resolution
// for TTL will be in seconds. The function is executed synchronously and
// it returns an error if some error was encountered during storage.
func (s *Noop) Store(m *message.Message) error {
	return nil
}

// Query performs a query and attempts to fetch last n messages where
// n is specified by limit argument. From and until times can also be specified
// for time-series retrieval.
func (s *Noop) Query(ssid message.Ssid, from, until time.Time, limit int) (message.Frame, error) {
	return nil, nil
}

// Close gracefully terminates the storage and ensures that every related
// resource is properly disposed.
func (s *Noop) Close() error {
	return nil
}
