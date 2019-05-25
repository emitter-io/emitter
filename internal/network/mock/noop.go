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

package mock

import (
	"net"
	"time"
)

// NewNoop returns a new noop.
func NewNoop() *Noop {
	return new(Noop)
}

// Noop is a fake connection.
type Noop struct{}

// Close closes the End.
func (e Noop) Close() (err error) {
	return
}

// LocalAddr gets the local address.
func (e Noop) LocalAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

// RemoteAddr gets the local address.
func (e Noop) RemoteAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

// Read implements the interface.
func (e Noop) Read(data []byte) (n int, err error) { return 0, nil }

// Write implements the interface.
func (e Noop) Write(data []byte) (n int, err error) { return len(data), nil }

// SetDeadline implements the interface.
func (e Noop) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline implements the interface.
func (e Noop) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline implements the interface.
func (e Noop) SetWriteDeadline(t time.Time) error { return nil }
