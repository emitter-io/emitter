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

package presence

import (
	"context"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/service"
	"github.com/kelindar/binary"
)

// Service represents a publish service.
type Service struct {
	context context.Context    // The context for the service.
	cancel  context.CancelFunc // The cancellation function.
	queue   chan *Notification // The channel for presence notifications.
	auth    service.Authorizer // The authorizer to use.
	pubsub  service.PubSub     // The pub/sub service to use.
	survey  service.Surveyor   // The surveyor to use.
	trie    *message.Trie      // The subscription matching trie.
}

// New creates a new publisher service.
func New(auth service.Authorizer, pubsub service.PubSub, survey service.Surveyor, trie *message.Trie) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{
		context: ctx,
		cancel:  cancel,
		auth:    auth,
		survey:  survey,
		pubsub:  pubsub,
		trie:    trie,
		queue:   make(chan *Notification, 100),
	}

	s.pollPresenceChange()
	return s
}

// notifyPresenceChange sends out an event to notify when a client is subscribed/unsubscribed.
func (s *Service) pollPresenceChange() {
	go func() {
		for {
			select {
			case <-s.context.Done():
				return
			case notif := <-s.queue:
				s.send(notif)
			}
		}
	}()
}

// send sends out an event to notify when a client is subscribed/unsubscribed.
func (s *Service) send(ev *Notification) {
	channel := []byte("emitter/presence/") // TODO: avoid allocation
	if encoded, ok := ev.Encode(); ok {
		s.pubsub.Publish(message.New(ev.Ssid, channel, encoded), ev.filter)
	}
}

// OnSurvey handles an incoming presence query.
func (s *Service) OnSurvey(queryType string, payload []byte) ([]byte, bool) {
	if queryType != "presence" {
		return nil, false
	}

	// Decode the request
	var target message.Ssid
	if err := binary.Unmarshal(payload, &target); err != nil {
		return nil, false
	}

	logging.LogTarget("query", queryType+" query received", target)

	// Send back the response
	presence, err := binary.Marshal(s.lookupPresence(target))
	return presence, err == nil
}

// lookupPresence performs a subscriptions lookup and returns a presence information.
func (s *Service) lookupPresence(ssid message.Ssid) []Info {
	resp := make([]Info, 0, 4)
	for _, subscriber := range s.trie.Lookup(ssid, nil) {
		if conn, ok := subscriber.(service.Conn); ok {
			resp = append(resp, Info{
				ID:       conn.ID(),
				Username: conn.Username(),
			})
		}
	}
	return resp
}

// Close closes gracefully the service.,
func (s *Service) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

// ------------------------------------------------------------------------------------

func (s *Service) getClusterPresence(ssid message.Ssid) []Info {
	who := make([]Info, 0, 4)
	if req, err := binary.Marshal(ssid); err == nil {
		if awaiter, err := s.survey.Query("presence", req); err == nil {

			// Wait for all presence updates to come back (or a deadline)
			for _, resp := range awaiter.Gather(1000 * time.Millisecond) {
				info := []Info{}
				if err := binary.Unmarshal(resp, &info); err == nil {
					//logging.LogTarget("query", "response gathered", info)
					who = append(who, info...)
				}
			}
		}
	}
	return who
}

func (s *Service) getLocalPresence(ssid message.Ssid) []Info {
	return s.lookupPresence(ssid)
}

func (s *Service) getAllPresence(ssid message.Ssid) []Info {
	return append(s.getLocalPresence(ssid), s.getClusterPresence(ssid)...)
}
