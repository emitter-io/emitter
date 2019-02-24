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
