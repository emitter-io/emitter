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
	"net"
	"testing"
	"time"

	"github.com/kelindar/rate"
	"github.com/stretchr/testify/assert"
)

func TestConn(t *testing.T) {
	conn := newConn(new(fakeConn))
	defer conn.Close()

	assert.Equal(t, 0, conn.Len())
	assert.Nil(t, conn.LocalAddr())
	assert.Nil(t, conn.RemoteAddr())
	assert.Nil(t, conn.SetDeadline(time.Now()))
	assert.Nil(t, conn.SetReadDeadline(time.Now()))
	assert.Nil(t, conn.SetWriteDeadline(time.Now()))

	conn.limit = rate.New(1, time.Millisecond)
	for i := 0; i < 100; i++ {
		_, err := conn.Write([]byte{1, 2, 3})
		assert.NoError(t, err)
	}
	time.Sleep(10 * time.Millisecond)
	_, err := conn.Write([]byte{1, 2, 3})
	assert.NoError(t, err)

}

// ------------------------------------------------------------------------------------

type fakeConn struct{}

func (m *fakeConn) Read(p []byte) (int, error) {
	return 0, nil
}

func (m *fakeConn) Write(p []byte) (int, error) {
	return 0, nil
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
