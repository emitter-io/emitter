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


package broker

import (
	"errors"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/broker/cluster"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
)

func Test_newQueryManager(t *testing.T) {
	s := new(Service)
	q := newQueryManager(s)

	assert.Equal(t, s, q.service)
	assert.Equal(t, uint32(0), q.next)
	assert.NotEqual(t, "", q.ID())
	assert.Equal(t, message.SubscriberDirect, q.Type())
}

func TestQuerySend_noSSID(t *testing.T) {
	q := newQueryManager(nil)

	err := q.Send(&message.Message{
		ID: message.NewID(message.Ssid{0, 0}),
	})
	assert.Error(t, errors.New("Invalid query received"), err)
}

func TestQuerySend_Response(t *testing.T) {
	q := newQueryManager(nil)

	err := q.Send(&message.Message{
		ID:      message.NewID(message.Ssid{0, 1, 2}),
		Channel: []byte("response"),
	})

	// There should be no awaiter, hence this should just pass
	assert.NoError(t, err)
}

/*func TestQuerySend_Request(t *testing.T) {
	q := newQueryManager(&Service{
		cluster: &cluster.Swarm{},
	})

	err := q.Send(&message.Message{
		ID:      message.NewID(message.Ssid{0, 1, 2}),
		Channel: []byte("request/12345/"),
	})

	assert.Equal(t, "No query handler found for request/12345/", err.Error())
}*/

func TestQuery_Query(t *testing.T) {
	q := newQueryManager(&Service{
		subscriptions: message.NewTrie(),
		cluster:       &cluster.Swarm{},
	})

	awaiter, err := q.Query("test", nil)
	assert.NoError(t, err)
	assert.NotNil(t, awaiter)

	result := awaiter.Gather(1 * time.Millisecond)
	assert.Empty(t, result)
}
