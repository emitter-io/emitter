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

package survey

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_New(t *testing.T) {
	q := New(nil, nil)

	assert.Nil(t, q.pubsub)
	assert.Nil(t, q.gossip)
	assert.Equal(t, uint32(0), q.next)
	assert.NotEqual(t, "", q.ID())
	assert.Equal(t, message.SubscriberDirect, q.Type())
}

func TestQuerySend_noSSID(t *testing.T) {
	q := New(nil, nil)

	err := q.Send(&message.Message{
		ID: message.NewID(message.Ssid{0, 0}),
	})
	assert.Error(t, errors.New("Invalid query received"), err)
}

func TestQuerySend_Response(t *testing.T) {
	q := New(nil, nil)

	err := q.Send(&message.Message{
		ID:      message.NewID(message.Ssid{0, 1, 2}),
		Channel: []byte("response"),
	})

	// There should be no awaiter, hence this should just pass
	assert.NoError(t, err)
}

func TestQuery(t *testing.T) {
	b1, out1 := newManager(1, 1)
	b2, out2 := newManager(2, 1)

	awaiter, err := b1.Query("test", nil)
	assert.NoError(t, err)
	assert.NotNil(t, awaiter)

	// Send the query to B2 and ourselves (should be ignored)
	query := <-out1
	assert.NoError(t, b1.Send(query))
	assert.NoError(t, b2.Send(query))

	// Send the response back to B1
	assert.NoError(t, b1.Send(<-out2))

	result := awaiter.Gather(1 * time.Millisecond)
	assert.Len(t, result, 1)
	assert.Equal(t, "hello from 2", string(result[0]))
}

func TestQuery_Timeout(t *testing.T) {
	b1, _ := newManager(1, 2)

	awaiter, err := b1.Query("test", nil)
	assert.NoError(t, err)
	assert.NotNil(t, awaiter)

	result := awaiter.Gather(1 * time.Millisecond)
	assert.Len(t, result, 0)
}

func TestQuery_NoPeers(t *testing.T) {
	b1, _ := newManager(1, 0)

	awaiter, err := b1.Query("test", nil)
	assert.NoError(t, err)
	assert.NotNil(t, awaiter)

	result := awaiter.Gather(1 * time.Millisecond)
	assert.Len(t, result, 0)
}

func Test_onRequest(t *testing.T) {
	g := new(gossiperMock)
	g.On("NumPeers").Return(2)
	g.On("ID").Return(uint64(5))
	q := New(nil, g)

	// Bad channel
	{
		err := q.onRequest(message.Ssid{1, 2}, "a/b/", nil)
		assert.Error(t, errors.New("Invalid query received"), err)
	}

	// No handler
	{
		err := q.onRequest(message.Ssid{1, 2}, "a/3/", nil)
		assert.Error(t, errors.New("Invalid query received"), err)
	}
}

func newManager(id, numPeers int) (*Surveyor, chan *message.Message) {
	b := new(pubsubMock)
	g := new(gossiperMock)
	s := new(surveyeeMock)
	s.On("OnSurvey", "test", mock.Anything).Return([]byte(fmt.Sprintf("hello from %d", id)), true)

	q := New(b, g)
	q.HandleFunc(s)
	out := make(chan *message.Message, 8)

	g.On("NumPeers").Return(numPeers)
	g.On("ID").Return(uint64(id))
	g.On("SendTo", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		out <- args.Get(1).(*message.Message)
	}).Return(nil)

	b.On("Subscribe", mock.Anything, mock.Anything).Return(true)
	b.On("Publish", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		out <- args.Get(0).(*message.Message)
	}).Return(int64(1))

	q.Start()
	return q, out
}
