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
	"fmt"
	"time"

	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/service"
)

var (
	_ service.Authorizer = new(Authorizer)
	_ service.Replicator = new(Replicator)
	_ service.PubSub     = new(PubSub)
	_ service.Conn       = new(Conn)
	_ service.Decryptor  = new(Decryptor)
	_ contract.Contract  = new(Contract)
	_ service.Surveyor   = new(Surveyor)
	_ service.Notifier   = new(Notifier)
)

// ------------------------------------------------------------------------------------

// Authorizer fake.
type Authorizer struct {
	Contract  uint32
	Target    string
	ExtraPerm uint8
	Success   bool
}

// Authorize provides a fake implementation.
func (f *Authorizer) Authorize(channel *security.Channel, perms uint8) (contract.Contract, security.Key, bool) {
	key := make(security.Key, 24)
	key.SetTarget(f.Target)
	key.SetPermissions(perms)
	if f.ExtraPerm > 0 {
		key.SetPermission(f.ExtraPerm, true)
	}

	key.SetContract(f.Contract)
	return &Contract{
		Invalid: !f.Success,
	}, key, f.Success
}

// ------------------------------------------------------------------------------------

// PubSub fake.
type PubSub struct {
	Trie *message.Trie
}

// Initializes the fake.
func (f *PubSub) initialize() {
	if f.Trie == nil {
		f.Trie = message.NewTrie()
	}
}

// Handle provides a fake implementation.
func (f *PubSub) Handle(_ string, _ service.Handler) {}

// Publish provides a fake implementation.
func (f *PubSub) Publish(m *message.Message, filter func(message.Subscriber) bool) (n int64) {
	f.initialize()
	size := m.Size()
	for _, subscriber := range f.Trie.Lookup(m.Ssid(), filter) {
		subscriber.Send(m)
		n += size
	}
	return size
}

// Subscribe provides a fake implementation.
func (f *PubSub) Subscribe(sub message.Subscriber, ev *event.Subscription) bool {
	f.initialize()
	f.Trie.Subscribe(ev.Ssid, sub)
	return true
}

// Unsubscribe provides a fake implementation.
func (f *PubSub) Unsubscribe(sub message.Subscriber, ev *event.Subscription) bool {
	f.initialize()
	f.Trie.Unsubscribe(ev.Ssid, sub)
	return true
}

// ------------------------------------------------------------------------------------

// Replicator fake.
type Replicator struct {
	data map[string]event.Event
}

// Initializes the fake.
func (f *Replicator) initialize() {
	if f.data == nil {
		f.data = make(map[string]event.Event)
	}
}

// Contains provides a fake implementation.
func (f *Replicator) Contains(ev event.Event) bool {
	f.initialize()
	_, ok := f.data[ev.Key()]
	return ok
}

// Notify provides a fake implementation.
func (f *Replicator) Notify(ev event.Event, enabled bool) {
	f.initialize()
	if enabled {
		f.data[ev.Key()] = ev
	} else {
		delete(f.data, ev.Key())
	}
}

// ------------------------------------------------------------------------------------

// Notifier fake.
type Notifier struct {
	Events []event.Subscription
}

// NotifySubscribe provides a fake implementation.
func (f *Notifier) NotifySubscribe(sub message.Subscriber, ev *event.Subscription) {
	f.Events = append(f.Events, *ev)
}

// NotifyUnsubscribe provides a fake implementation.
func (f *Notifier) NotifyUnsubscribe(sub message.Subscriber, ev *event.Subscription) {
	f.Events = append(f.Events, *ev)
}

// ------------------------------------------------------------------------------------

// Conn fake.
type Conn struct {
	ConnID    int
	Disabled  bool
	Outgoing  []message.Message
	Shortcuts map[string]string
}

// Initializes the fake.
func (f *Conn) initialize() {
	if f.Shortcuts == nil {
		f.Shortcuts = make(map[string]string)
	}
}

// Close provides a fake implementation.
func (f *Conn) Close() error {
	return nil
}

// ID provides a fake implementation.
func (f *Conn) ID() string {
	return fmt.Sprintf("%v", f.ConnID)
}

// Type provides a fake implementation.
func (f *Conn) Type() message.SubscriberType {
	return message.SubscriberDirect
}

// Send provides a fake implementation.
func (f *Conn) Send(m *message.Message) error {
	f.Outgoing = append(f.Outgoing, *m)
	return nil
}

// CanSubscribe provides a fake implementation.
func (f *Conn) CanSubscribe(message.Ssid, []byte) bool {
	return !f.Disabled
}

// CanUnsubscribe provides a fake implementation.
func (f *Conn) CanUnsubscribe(message.Ssid, []byte) bool {
	return !f.Disabled
}

// LocalID provides a fake implementation.
func (f *Conn) LocalID() security.ID {
	return security.ID(f.ConnID)
}

// Username provides a fake implementation.
func (f *Conn) Username() string {
	return fmt.Sprintf("user of %v", f.ConnID)
}

// Track provides a fake implementation.
func (f *Conn) Track(contract.Contract) {

}

// Links provides a fake implementation.
func (f *Conn) Links() map[string]string {
	f.initialize()
	return f.Shortcuts
}

// GetLink provides a fake implementation.
func (f *Conn) GetLink(input []byte) []byte {
	f.initialize()
	if v, ok := f.Shortcuts[string(input)]; ok {
		input = []byte(v)
	}
	return input
}

// AddLink provides a fake implementation.
func (f *Conn) AddLink(alias string, channel *security.Channel) {
	f.initialize()
	f.Shortcuts[alias] = channel.String()
}

// ------------------------------------------------------------------------------------

// Decryptor fake.
type Decryptor struct {
	Contract    uint32
	Permissions uint8
	Target      string
}

// DecryptKey provides a fake implementation.
func (f *Decryptor) DecryptKey(k string) (security.Key, error) {
	key := make(security.Key, 24)
	key.SetTarget(f.Target)
	key.SetPermissions(f.Permissions)
	key.SetContract(f.Contract)
	return key, nil
}

// ------------------------------------------------------------------------------------

// Contract fake.
type Contract struct {
	Invalid bool
}

// Validate validates the contract data against a key.
func (f *Contract) Validate(key security.Key) bool {
	return !f.Invalid
}

// Stats gets the usage statistics.
func (f *Contract) Stats() usage.Meter {
	return usage.NewNoop().Get(1)
}

// ------------------------------------------------------------------------------------

// Surveyor fake.
type Surveyor struct {
	Resp [][]byte
	Err  error
}

// Query provides a fake implementation.
func (f *Surveyor) Query(string, []byte) (message.Awaiter, error) {
	return &awaiter{f.Resp}, f.Err
}

type awaiter struct {
	r [][]byte
}

func (a *awaiter) Gather(timeout time.Duration) [][]byte {
	return a.r
}
