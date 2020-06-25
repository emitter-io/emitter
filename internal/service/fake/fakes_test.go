/**********************************************************************************
* Copyright (c) 2009-2020 Misakai Ltd.
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

package fake

import (
	"testing"

	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizer(t *testing.T) {
	f := &Authorizer{
		Success:   true,
		ExtraPerm: security.AllowExtend,
	}

	c, _, ok := f.Authorize(nil, 1)
	assert.True(t, ok)
	assert.True(t, c.Validate(nil))
	assert.NotNil(t, c.Stats())
}

func TestPubSub(t *testing.T) {
	f := new(PubSub)
	ev := &event.Subscription{
		Ssid: message.Ssid{1, 2},
	}

	conn := new(Conn)
	assert.True(t, f.Unsubscribe(conn, ev))
	assert.True(t, f.Subscribe(conn, ev))

	n := f.Publish(message.New(
		message.Ssid{1, 2},
		[]byte("a/"),
		[]byte("---"),
	), nil)
	assert.Equal(t, int64(3), n)
	assert.Len(t, conn.Outgoing, 1)

	f.Handle("", nil)
}

func TestReplicator(t *testing.T) {
	f := new(Replicator)
	ev := event.Ban("abc")

	f.Notify(&ev, true)
	assert.True(t, f.Contains(&ev))
	f.Notify(&ev, false)
	assert.False(t, f.Contains(&ev))
}

func TestNotifier(t *testing.T) {
	f := new(Notifier)
	ev := &event.Subscription{
		Ssid: message.Ssid{1, 2},
	}

	f.NotifySubscribe(nil, ev)
	f.NotifyUnsubscribe(nil, ev)
	assert.Len(t, f.Events, 2)
}

func TestConn(t *testing.T) {
	f := &Conn{ConnID: 1}
	f.Track(nil)

	assert.Equal(t, security.ID(1), f.LocalID())
	assert.Equal(t, "1", f.ID())
	assert.Equal(t, "user of 1", f.Username())
	assert.Equal(t, message.SubscriberDirect, f.Type())
	assert.NoError(t, f.Close())
	assert.True(t, f.CanSubscribe(nil, nil))
	assert.True(t, f.CanUnsubscribe(nil, nil))

	f.AddLink("a", &security.Channel{})
	assert.NotNil(t, f.GetLink([]byte("a")))
	assert.Len(t, f.Links(), 1)
}

func TestDecryptor(t *testing.T) {
	f := &Decryptor{
		Contract:    1,
		Permissions: security.AllowRead,
		Target:      "a/b/c/",
	}

	k, err := f.DecryptKey("")
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), k.Contract())
}

func TestSurvey(t *testing.T) {
	f := &Surveyor{
		Resp: [][]byte{[]byte("hi")},
	}

	r, err := f.Query("", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Gather(0)))
}
