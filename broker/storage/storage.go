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
	"io"
	"time"
)

// Storage represents a message storage contract that message storage provides
// must fulfill.
type Storage interface {
	io.Closer

	// Configure configures the storage. The config parameter provided is
	// loosely typed, since various storage mechanisms will require different
	// configurations.
	Configure(config map[string]interface{}) error

	// Store is used to store a message, the SSID provided must be a full SSID
	// SSID, where first element should be a contract ID. The time resolution
	// for TTL will be in seconds. The function is executed synchronously and
	// it returns an error if some error was encountered during storage.
	Store(ssid []uint32, payload []byte, ttl time.Duration) error

	// QueryLast performs a query and attempts to fetch last n messages where
	// n is specified by limit argument. It returns a channel which will be
	// ranged over to retrieve messages asynchronously.
	QueryLast(ssid []uint32, limit int) (<-chan []byte, error)
}
