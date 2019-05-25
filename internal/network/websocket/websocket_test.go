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

package websocket

import (
	"bytes"
	"io"
	"net"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type writer bytes.Buffer

func (w *writer) Close() error                         { return nil }
func (w *writer) Write(data []byte) (n int, err error) { return ((*bytes.Buffer)(w)).Write(data) }

type conn struct {
	read  []byte
	write *writer
}

func (c *conn) NextReader() (messageType int, r io.Reader, err error) {
	messageType = websocket.BinaryMessage
	r = bytes.NewBuffer(c.read)
	if c.read == nil {
		err = io.EOF
	}
	return
}

func (c *conn) NextWriter(messageType int) (w io.WriteCloser, err error) {
	w = c.write
	if c.write == nil {
		err = io.EOF
	}

	return
}
func (c *conn) Close() error                       { return nil }
func (c *conn) LocalAddr() net.Addr                { return &net.IPAddr{} }
func (c *conn) RemoteAddr() net.Addr               { return &net.IPAddr{} }
func (c *conn) SetReadDeadline(t time.Time) error  { return nil }
func (c *conn) SetWriteDeadline(t time.Time) error { return nil }

func TestTryUpgradeNil(t *testing.T) {
	_, ok := TryUpgrade(nil, nil)
	assert.Equal(t, false, ok)
}

func TestTryUpgrade(t *testing.T) {
	//httptest.NewServer(handler)
	r := httptest.NewRequest("GET", "http://127.0.0.1/", bytes.NewBuffer([]byte{}))
	r.Header.Set("Connection", "upgrade")
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Sec-WebSocket-Extensions", "permessage-deflate; client_max_window_bits")
	r.Header.Set("Sec-WebSocket-Key", "D1icfJz+khA9kj5/14dRXQ==")
	r.Header.Set("Sec-WebSocket-Protocol", "mqttv3.1")
	r.Header.Set("Sec-WebSocket-Version", "13")

	w := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		TryUpgrade(w, r)
	})

	// TODO: need to have a hijackable response writer to test properly
	//ws, ok := TryUpgrade(w, r)
	//assert.NotNil(t, ws)
	//assert.True(t, ok)
}

func TestRead_EOF(t *testing.T) {
	c := newConn(new(conn))

	_, err := c.Read([]byte{})
	assert.Error(t, io.EOF, err)
}

func TestRead(t *testing.T) {
	message := []byte("hello world")
	c := &websocketTransport{
		socket: &conn{
			read: message,
		},
		closing: make(chan bool),
	}

	buffer := make([]byte, 64)
	n, err := c.Read(buffer)
	assert.NoError(t, err)
	assert.Equal(t, message, buffer[:n])
}

func TestWrite(t *testing.T) {
	message := []byte("hello world")
	buffer := new(bytes.Buffer)
	c := &websocketTransport{
		socket: &conn{
			write: (*writer)(buffer),
		},
		closing: make(chan bool),
	}

	_, err := c.Write(message)
	assert.NoError(t, err)
	assert.Equal(t, message, buffer.Bytes())
}

func TestMisc(t *testing.T) {
	c := &websocketTransport{
		socket:  &conn{},
		closing: make(chan bool),
	}

	err := c.Close()
	assert.NoError(t, err)

	err = c.SetDeadline(time.Now())
	assert.NoError(t, err)

	err = c.SetReadDeadline(time.Now())
	assert.NoError(t, err)

	err = c.SetWriteDeadline(time.Now())
	assert.NoError(t, err)

	addr1 := c.LocalAddr()
	assert.Equal(t, "", addr1.String())

	addr2 := c.RemoteAddr()
	assert.Equal(t, "", addr2.String())
}
