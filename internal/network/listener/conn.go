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

package listener

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/emitter-io/emitter/internal/async"
	"github.com/kelindar/rate"
)

// Conn wraps a net.Conn and provides transparent sniffing of connection data.
type Conn struct {
	sync.RWMutex
	socket net.Conn           // The underlying network connection.
	writer bytes.Buffer       // The buffered write queue.
	reader sniffer            // The reader which performs sniffing.
	limit  *rate.Limiter      // The write rate limiter.
	cancel context.CancelFunc // The cancellation function for the force flush.
}

// NewConn creates a new sniffed connection.
func newConn(c net.Conn, writeRate int) *Conn {
	if writeRate <= 0 || writeRate > 1000 {
		writeRate = 60
	}

	conn := &Conn{
		socket: c,
		reader: sniffer{source: c},
		limit:  rate.New(writeRate, time.Second),
	}

	// TODO: see if we can get rid of this goroutine per connection
	conn.cancel = async.Repeat(context.Background(), 1*time.Second, func() { conn.Flush() })
	return conn
}

// Read reads the block of data from the underlying buffer.
func (m *Conn) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

// Write writes the block of data into the underlying buffer.
func (m *Conn) Write(p []byte) (int, error) {

	// If we have reached the limit we can possibly write, queue up the packet.
	if m.limit.Limit() {
		return m.enqueue(p)
	}

	// If we have something in the buffer, flush everything.
	if m.Len() > 0 {
		m.enqueue(p)
		return m.Flush()
	}

	// Nothing in the buffer and we're not rate-limited, just write to the socket.
	return m.socket.Write(p)
}

// Close closes the connection. Any blocked Read or Write operations will be unblocked
// and return errors.
func (m *Conn) Close() error {
	m.cancel()
	return m.socket.Close()
}

// Flush flushes the underlying buffer by writing into the underlying connection.
func (m *Conn) Flush() (n int, err error) {
	if m.Len() == 0 {
		return 0, nil
	}

	// Flush everything and reset the buffer
	m.Lock()
	n, err = m.socket.Write(m.writer.Bytes())
	m.writer.Reset()
	m.Unlock()
	return
}

// LocalAddr returns the local network address.
func (m *Conn) LocalAddr() net.Addr {
	return m.socket.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (m *Conn) RemoteAddr() net.Addr {
	return m.socket.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
func (m *Conn) SetDeadline(t time.Time) error {
	return m.socket.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
func (m *Conn) SetReadDeadline(t time.Time) error {
	return m.socket.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
func (m *Conn) SetWriteDeadline(t time.Time) error {
	return m.socket.SetWriteDeadline(t)
}

// Len returns the pending buffer size.
func (m *Conn) Len() (n int) {
	m.RLock()
	n = int(m.writer.Len())
	m.RUnlock()
	return
}

func (m *Conn) enqueue(p []byte) (n int, err error) {
	m.Lock()
	n, err = m.writer.Write(p)
	m.Unlock()
	return
}

func (m *Conn) startSniffing() io.Reader {
	m.reader.reset(true)
	return &m.reader
}

func (m *Conn) doneSniffing() {
	m.reader.reset(false)
}

// ------------------------------------------------------------------------------------

// Sniffer represents a io.Reader which can peek incoming bytes and reset back to normal.
type sniffer struct {
	source     io.Reader
	buffer     bytes.Buffer
	bufferRead int
	bufferSize int
	sniffing   bool
	lastErr    error
}

// Read reads data from the buffer.
func (s *sniffer) Read(p []byte) (int, error) {
	if s.bufferSize > s.bufferRead {
		bn := copy(p, s.buffer.Bytes()[s.bufferRead:s.bufferSize])
		s.bufferRead += bn
		return bn, s.lastErr
	} else if !s.sniffing && s.buffer.Cap() != 0 {
		s.buffer = bytes.Buffer{}
	}

	sn, sErr := s.source.Read(p)
	if sn > 0 && s.sniffing {
		s.lastErr = sErr
		if wn, wErr := s.buffer.Write(p[:sn]); wErr != nil {
			return wn, wErr
		}
	}
	return sn, sErr
}

// Reset resets the buffer.
func (s *sniffer) reset(snif bool) {
	s.sniffing = snif
	s.bufferRead = 0
	s.bufferSize = s.buffer.Len()
}
