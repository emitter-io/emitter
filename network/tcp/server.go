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

package tcp

import (
	"errors"
	"net"
	"sync"
	"time"

	"fmt"
	"github.com/emitter-io/emitter/logging"
)

// Handler is a callback which gets called when a new connection is accepted.
type Handler func(c net.Conn)

// ErrServerClosed occurs wehen a tcp server is closed.
var ErrServerClosed = errors.New("tcp: Server closed")

// Server represents a TCP server.
type Server struct {
	sync.Mutex

	Handler Handler
	Closing chan bool
}

// ServeAsync creates a TCP listener and starts the server.
func ServeAsync(port int, closing chan bool, handler Handler) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := new(Server)
	server.Closing = closing
	server.Handler = handler
	go server.Serve(l)
	return nil
}

// Serve accepts the connections and fires the callback
func (s *Server) Serve(l net.Listener) error {
	defer l.Close()

	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-s.closing():
				return ErrServerClosed
			default:
			}

			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				logging.LogError("tcp", "accept error", err)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		tempDelay = 0
		go s.Handler(conn)
	}
}

// Close immediately closes all active listeners.
func (s *Server) Close() error {
	s.Lock()
	defer s.Unlock()
	s.close()
	return nil
}

// Closing gets the closing channel in a thread-safe manner.
func (s *Server) closing() <-chan bool {
	s.Lock()
	defer s.Unlock()
	return s.getClosing()
}

// Closing gets the closing channel in a non thread-safe manner.
func (s *Server) getClosing() chan bool {
	if s.Closing == nil {
		s.Closing = make(chan bool)
	}
	return s.Closing
}

func (s *Server) close() {
	ch := s.getClosing()
	select {
	case <-ch:
		// Already closed. Don't close again.
	default:
		// Safe to close here. We're the only closer, guarded
		// by s.mu.
		close(ch)
	}
}
