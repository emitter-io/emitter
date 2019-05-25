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

package monitor

import (
	"io"
	"time"

	"github.com/emitter-io/config"
)

var defaultInterval = 5 * time.Second

// Storage represents a message storage contract that message storage provides
// must fulfill.
type Storage interface {
	config.Provider
	io.Closer
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

// Close gracefully terminates the storage and ensures that every related
// resource is properly disposed.
func (s *Noop) Close() error {
	return nil
}
