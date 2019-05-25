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
	"io"
	"net"
	"time"
)

// End is one 'end' of a simulated connection.
type End struct {
	Reader *io.PipeReader
	Writer *io.PipeWriter
}

// Close closes the End.
func (e End) Close() (err error) {
	if err = e.Writer.Close(); err == nil {
		err = e.Reader.Close()
	}
	return
}

// LocalAddr gets the local address.
func (e End) LocalAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

// RemoteAddr gets the local address.
func (e End) RemoteAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

// Read implements the interface.
func (e End) Read(data []byte) (n int, err error) { return e.Reader.Read(data) }

// Write implements the interface.
func (e End) Write(data []byte) (n int, err error) { return e.Writer.Write(data) }

// SetDeadline implements the interface.
func (e End) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline implements the interface.
func (e End) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline implements the interface.
func (e End) SetWriteDeadline(t time.Time) error { return nil }
