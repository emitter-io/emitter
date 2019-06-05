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
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Server represents a server which can serve requests.
type Server interface {
	Serve(listener net.Listener)
}

// Matcher matches a connection based on its content.
type Matcher func(io.Reader) bool

// ErrorHandler handles an error and notifies the listener on whether
// it should continue serving.
type ErrorHandler func(error) bool

var _ net.Error = ErrNotMatched{}

// ErrNotMatched is returned whenever a connection is not matched by any of
// the matchers registered in the multiplexer.
type ErrNotMatched struct {
	c net.Conn
}

func (e ErrNotMatched) Error() string {
	return fmt.Sprintf("Unable to match connection %v", e.c.RemoteAddr())
}

// Temporary implements the net.Error interface.
func (e ErrNotMatched) Temporary() bool { return true }

// Timeout implements the net.Error interface.
func (e ErrNotMatched) Timeout() bool { return false }

type errListenerClosed string

func (e errListenerClosed) Error() string   { return string(e) }
func (e errListenerClosed) Temporary() bool { return false }
func (e errListenerClosed) Timeout() bool   { return false }

// ErrListenerClosed is returned from muxListener.Accept when the underlying
// listener is closed.
var ErrListenerClosed = errListenerClosed("mux: listener closed")

// for readability of readTimeout
var noTimeout time.Duration

// Config represents the configuration of the listener.
type Config struct {
	TLS       *tls.Config // The TLS/SSL configuration.
	FlushRate int         // The maximum flush rate (QPS) per connection.
}

// New announces on the local network address laddr. The syntax of laddr is
// "host:port", like "127.0.0.1:8080". If host is omitted, as in ":8080",
// New listens on all available interfaces instead of just the interface
// with the given host address. Listening on a hostname is not recommended
// because this creates a socket for at most one of its IP addresses.
func New(address string, config Config) (*Listener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	// If we have a TLS configuration provided, wrap the listener in TLS
	if config.TLS != nil {
		l = tls.NewListener(l, config.TLS)
	}

	return &Listener{
		root:         l,
		bufferSize:   1024,
		errorHandler: func(_ error) bool { return true },
		closing:      make(chan struct{}),
		readTimeout:  noTimeout,
		config:       config,
	}, nil
}

type processor struct {
	matchers []Matcher
	listen   muxListener
}

// Listener represents a listener used for multiplexing protocols.
type Listener struct {
	root         net.Listener
	bufferSize   int
	errorHandler ErrorHandler
	closing      chan struct{}
	matchers     []processor
	readTimeout  time.Duration
	config       Config
}

// Accept waits for and returns the next connection to the listener.
func (m *Listener) Accept() (net.Conn, error) {
	return m.root.Accept()
}

// ServeAsync adds a protocol based on the matcher and serves it.
func (m *Listener) ServeAsync(matcher Matcher, serve func(l net.Listener) error) {
	l := m.Match(matcher)
	go serve(l)
}

// Match returns a net.Listener that sees (i.e., accepts) only
// the connections matched by at least one of the matcher.
func (m *Listener) Match(matchers ...Matcher) net.Listener {
	ml := muxListener{
		Listener:    m.root,
		connections: make(chan net.Conn, m.bufferSize),
	}
	m.matchers = append(m.matchers, processor{matchers: matchers, listen: ml})
	return ml
}

// SetReadTimeout sets a timeout for the read of matchers.
func (m *Listener) SetReadTimeout(t time.Duration) {
	m.readTimeout = t
}

// Serve starts multiplexing the listener.
func (m *Listener) Serve() error {
	var wg sync.WaitGroup

	defer func() {
		close(m.closing)
		wg.Wait()

		for _, sl := range m.matchers {
			close(sl.listen.connections)
			// Drain the connections enqueued for the listener.
			for c := range sl.listen.connections {
				_ = c.Close()
			}
		}
	}()

	for {
		c, err := m.root.Accept()
		if err != nil {
			if !m.handleErr(err) {
				return err
			}
			continue
		}

		wg.Add(1)
		go m.serve(c, m.closing, &wg)
	}
}

func (m *Listener) serve(c net.Conn, donec <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	muc := newConn(c, m.config.FlushRate)
	if m.readTimeout > noTimeout {
		_ = c.SetReadDeadline(time.Now().Add(m.readTimeout))
	}
	for _, sl := range m.matchers {
		for _, processor := range sl.matchers {
			matched := processor(muc.startSniffing())
			if matched {
				muc.doneSniffing()
				if m.readTimeout > noTimeout {
					_ = c.SetReadDeadline(time.Time{})
				}
				select {
				case sl.listen.connections <- muc:
				case <-donec:
					_ = c.Close()
				}
				return
			}
		}
	}

	_ = c.Close()
	err := ErrNotMatched{c: c}
	if !m.handleErr(err) {
		_ = m.root.Close()
	}
}

// HandleError registers an error handler that handles listener errors.
func (m *Listener) HandleError(h ErrorHandler) {
	m.errorHandler = h
}

func (m *Listener) handleErr(err error) bool {
	if !m.errorHandler(err) {
		return false
	}

	if ne, ok := err.(net.Error); ok {
		return ne.Temporary()
	}

	return false
}

// Close closes the listener
func (m *Listener) Close() error {
	return m.root.Close()
}

// Addr returns the listener's network address.
func (m *Listener) Addr() net.Addr {
	return m.root.Addr()
}

// ------------------------------------------------------------------------------------

type muxListener struct {
	net.Listener
	connections chan net.Conn
}

func (l muxListener) Accept() (net.Conn, error) {
	c, ok := <-l.connections
	if !ok {
		return nil, ErrListenerClosed
	}
	return c, nil
}
