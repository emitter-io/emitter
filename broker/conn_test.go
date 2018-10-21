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

package broker

import (
	"io/ioutil"
	"testing"

	"github.com/emitter-io/emitter/message"
	netmock "github.com/emitter-io/emitter/network/mock"
	"github.com/emitter-io/emitter/security"
	"github.com/emitter-io/stats"
	"github.com/stretchr/testify/assert"
)

func newTestConn() (pipe *netmock.Conn, conn *Conn) {
	license, _ := security.ParseLicense(testLicense)
	s := &Service{
		subscriptions: message.NewTrie(),
		License:       license,
		measurer:      stats.NewNoop(),
	}

	pipe = netmock.NewConn()
	conn = s.newConn(pipe.Client)
	return
}

func TestNotifyError(t *testing.T) {
	pipe, conn := newTestConn()
	assert.NotNil(t, pipe)

	go func() {
		conn.notifyError(ErrUnauthorized)
		conn.Close()
	}()

	b, err := ioutil.ReadAll(pipe.Server)
	assert.Contains(t, string(b), ErrUnauthorized.Message)
	assert.NoError(t, err)
}
