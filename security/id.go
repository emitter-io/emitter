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

package security

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
	"sync/atomic"
	"time"
)

// ID represents a process-wide unique ID.
type ID uint64

// next is the next identifier. We seed it with the time in seconds
// to avoid collisions of ids between process restarts.
var next = uint64(
	time.Now().Sub(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)).Seconds(),
)

// NewID generates a new, process-wide unique ID.
func NewID() ID {
	return ID(atomic.AddUint64(&next, 1))
}

// String converts the ID to a string representation.
func (id ID) String() string {
	buf := make([]byte, 10) // This will never be more than 9 bytes.
	l := binary.PutUvarint(buf, uint64(id))
	return strings.ToUpper(hex.EncodeToString(buf[:l]))
}
