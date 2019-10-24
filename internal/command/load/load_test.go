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

package load

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/network/mqtt"
	cli "github.com/jawher/mow.cli"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	dial = func(string, string) (net.Conn, error) {
		conn := new(fakeConn)

		// Push connack
		connack := new(mqtt.Connack)
		connack.EncodeTo(&conn.read)

		// Push suback
		suback := new(mqtt.Suback)
		suback.EncodeTo(&conn.read)
		return conn, nil
	}

	runCommand("emitter", "test", "key")
}

func Test_newConn_Error(t *testing.T) {
	dial = func(string, string) (net.Conn, error) {
		return nil, io.EOF
	}

	conn, _ := newConn("127.0.0.1:8080", "", "")
	assert.Nil(t, conn)
	assert.NotPanics(t, func() {
		runCommand("emitter", "test", "key")
	})
}

func Test_newConn(t *testing.T) {
	dial = func(string, string) (net.Conn, error) {
		return new(fakeConn), nil
	}

	conn, _ := newConn("127.0.0.1:8080", "", "")
	assert.NotNil(t, conn)

	_, err := conn.ReadByte()
	assert.Equal(t, io.EOF, err)
}

func Test_newMessage(t *testing.T) {
	msg := newMessage("a/", -1)
	assert.Equal(t, "a/", string(msg.Topic))
}

func runCommand(args ...string) {
	app := cli.App("emitter", "")
	app.Command("test", "", Run)
	app.Run(args)
}

// ------------------------------------------------------------------------------------

type fakeConn struct {
	read  bytes.Buffer
	write int
}

func (m *fakeConn) Read(p []byte) (int, error) {
	return m.read.Read(p)
}

func (m *fakeConn) Write(p []byte) (int, error) {
	if m.write++; m.write < 100 {
		return len(p), nil
	}

	return 0, io.EOF
}

func (m *fakeConn) Close() error {
	return nil
}

func (m *fakeConn) LocalAddr() net.Addr {
	return nil
}

func (m *fakeConn) RemoteAddr() net.Addr {
	return nil
}

func (m *fakeConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *fakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *fakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}
