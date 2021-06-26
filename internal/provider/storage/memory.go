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
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/emitter-io/emitter/internal/async"
	"github.com/emitter-io/emitter/internal/service"
)

// InMemory implements Storage contract.
var _ Storage = new(InMemory)

// InMemory represents a storage which does nothing.
type InMemory struct {
	SSD // Badger with in-memory mode
}

// NewInMemory creates a new in-memory storage.
func NewInMemory(survey service.Surveyor) *InMemory {
	return &InMemory{SSD{
		survey: survey,
	}}
}

// Name returns the name of the provider.
func (s *InMemory) Name() string {
	return "inmemory"
}

// Configure configures the storage. The config parameter provided is
// loosely typed, since various storage mechanisms will require different
// configurations.
func (s *InMemory) Configure(config map[string]interface{}) error {
	opts := badger.DefaultOptions("")
	opts.SyncWrites = true
	opts.InMemory = true

	// Attempt to open the database
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}

	// Setup the database and start GC
	s.db = db
	s.retain = configUint32(config, "retain", defaultRetain)
	s.cancel = async.Repeat(context.Background(), 30*time.Minute, s.GC)
	return err
}
