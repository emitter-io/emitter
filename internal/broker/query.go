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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/weaveworks/mesh"
)

const (
	idSystem = uint32(0)
	idQuery  = uint32(3939663052)
)

// Surveyee handles the surveys.
type Surveyee interface {
	OnSurvey(queryType string, request []byte) (response []byte, ok bool)
}

// QueryManager represents a request-response manager.
type QueryManager struct {
	service  *Service    // The service to use.
	luid     security.ID // The locally unique id of the manager.
	next     uint32      // The next available query identifier.
	awaiters *sync.Map   // The map of the awaiters.
	handlers []Surveyee  // The handlers array.
}

// newQueryManager creates a new request-response manager.
func newQueryManager(s *Service) *QueryManager {
	return &QueryManager{
		service:  s,
		luid:     security.NewID(),
		next:     0,
		awaiters: new(sync.Map),
		handlers: make([]Surveyee, 0),
	}
}

// Start subscribes the manager to the query channel.
func (c *QueryManager) Start() {
	ssid := message.Ssid{idSystem, idQuery}
	if ok := c.service.onSubscribe(ssid, c); ok {
		c.service.cluster.NotifySubscribe(c.luid, ssid)
	}
}

// HandleFunc adds a handler for a query.
func (c *QueryManager) HandleFunc(surveyees ...Surveyee) {
	for _, h := range surveyees {
		c.handlers = append(c.handlers, h)
	}
}

// ID returns the unique identifier of the subsriber.
func (c *QueryManager) ID() string {
	return c.luid.String()
}

// Type returns the type of the subscriber
func (c *QueryManager) Type() message.SubscriberType {
	return message.SubscriberDirect
}

// Send occurs when we have received a message.
func (c *QueryManager) Send(m *message.Message) error {
	ssid := m.Ssid()
	if len(ssid) != 3 {
		return errors.New("Invalid query received")
	}

	switch string(m.Channel) {
	case "response":
		// We received a response, find the awaiter and forward a message to it
		return c.onResponse(ssid[2], m.Payload)

	default:
		// We received a request, need to handle that by calling the appropriate handler
		return c.onRequest(ssid, string(m.Channel), m.Payload)
	}
}

// onRequest handles an incoming request
func (c *QueryManager) onResponse(id uint32, payload []byte) error {
	if awaiter, ok := c.awaiters.Load(id); ok {
		awaiter.(*queryAwaiter).receive <- payload
	}
	return nil
}

// onRequest handles an incoming request
func (c *QueryManager) onRequest(ssid message.Ssid, channel string, payload []byte) error {
	// Get the query and reply node
	ch := strings.Split(channel, "/")
	query := ch[0]
	reply, err := strconv.ParseInt(ch[1], 10, 64)
	if err != nil {
		return err
	}

	// Do not answer our own requests
	replyAddr := mesh.PeerName(reply)
	if c.service.cluster.ID() == uint64(replyAddr) {
		return nil
	}

	// Get the peer to reply to
	peer := c.service.cluster.FindPeer(replyAddr)
	if !peer.IsActive() {
		return errors.New("unable to reply to a request, peer is not active")
	}

	// Go through all the handlers and execute the first matching one
	for _, surveyee := range c.handlers {
		if response, ok := surveyee.OnSurvey(query, payload); ok {
			return peer.Send(message.New(ssid, []byte("response"), response))
		}
	}

	return errors.New("no query handler found for " + channel)
}

// Query issues a cluster-wide request.
func (c *QueryManager) Query(query string, payload []byte) (message.Awaiter, error) {

	// Create an awaiter
	// TODO: replace the max with the total number of cluster nodes
	numPeers := c.service.NumPeers()
	awaiter := &queryAwaiter{
		id:      atomic.AddUint32(&c.next, 1),
		receive: make(chan []byte, numPeers),
		maximum: numPeers,
		manager: c,
	}

	// Store an awaiter
	c.awaiters.Store(awaiter.id, awaiter)

	// Prepare a channel with the reply-to address
	channel := fmt.Sprintf("%v/%v", query, c.service.LocalName())

	// Publish the query as a message
	c.service.publish(message.New(
		message.Ssid{idSystem, idQuery, awaiter.id},
		[]byte(channel),
		payload,
	), "")
	return awaiter, nil
}

// queryAwaiter represents an asynchronously awaiting response channel.
type queryAwaiter struct {
	id      uint32        // The identifier of the query.
	maximum int           // The maximum number of responses to wait for.
	receive chan []byte   // The receive channel to use.
	manager *QueryManager // The query manager used.
}

// Gather awaits for the responses to be received, blocking until we're done.
func (a *queryAwaiter) Gather(timeout time.Duration) (r [][]byte) {
	defer func() { a.manager.awaiters.Delete(a.id) }()
	r = make([][]byte, 0, 4)
	t := time.After(timeout)
	c := a.maximum

	// If there's no peers, no need to receive anything
	if c == 0 {
		return
	}

	for {
		select {
		case msg := <-a.receive:
			r = append(r, msg)
			c-- // Decrement the counter
			if c == 0 {
				return // We got all the responses we needed
			}

		case <-t:
			return // We timed out
		}
	}
}
