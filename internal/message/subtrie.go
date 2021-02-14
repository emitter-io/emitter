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

package message

import (
	"sync"
	"time"
)

type node struct {
	word     uint32
	subs     Subscribers
	parent   *node
	children map[uint32]*node
}

func (n *node) orphan() {
	if n.parent == nil {
		return
	}

	delete(n.parent.children, n.word)
	if n.parent.subs.Size() == 0 && len(n.parent.children) == 0 {
		n.parent.orphan()
	}
}

// Trie represents an efficient collection of subscriptions with lookup capability.
type Trie struct {
	sync.RWMutex
	root  *node // The root node of the tree.
	count int   // Number of subscriptions in the trie.
	lookup func(Ssid, *Subscribers, *node, func(s Subscriber) bool)
}

// newTrie creates a new trie without a lookup function
func newTrie() *Trie {
	return &Trie{
		root: &node{
			subs:     newSubscribers(),
			children: make(map[uint32]*node),
		},
	}
}

// NewTrie creates a new subscriptions matcher using standard emitter strategy.
func NewTrie() *Trie {
	t := newTrie()
	t.lookup = t.lookupEmitter
	return t
}

// NewTrieMQTT creates a new subscriptions matcher using standard MQTT strategy.
func NewTrieMQTT() *Trie {
	t := newTrie()
	t.lookup = t.lookupMqtt
	return t
}

// Count returns the number of subscriptions.
func (t *Trie) Count() int {
	t.RLock()
	defer t.RUnlock()
	return t.count
}

// Subscribe adds the Subscriber to the topic and returns a Subscription.
func (t *Trie) Subscribe(ssid Ssid, sub Subscriber) *Subscription {
	t.Lock()
	curr := t.root
	for _, word := range ssid {
		child, ok := curr.children[word]
		if !ok {
			child = &node{
				word:     word,
				subs:     newSubscribers(),
				parent:   curr,
				children: make(map[uint32]*node),
			}
			curr.children[word] = child
		}
		curr = child
	}

	// Add unique and count
	if ok := curr.subs.AddUnique(sub); ok {
		t.count++
	}

	t.Unlock()
	return &Subscription{Ssid: ssid, Subscriber: sub}
}

// Unsubscribe removes the Subscription.
func (t *Trie) Unsubscribe(ssid Ssid, subscriber Subscriber) {
	t.Lock()
	curr := t.root
	for _, word := range ssid {
		child, ok := curr.children[word]
		if !ok {
			// Subscription doesn't exist.
			t.Unlock()
			return
		}
		curr = child
	}

	// Remove the subscriber and decrement the counter
	if ok := curr.subs.Remove(subscriber); ok {
		t.count--
	}

	// Remove orphans
	if curr.subs.Size() == 0 && len(curr.children) == 0 {
		curr.orphan()
	}
	t.Unlock()
}

// Lookup returns the Subscribers for the given topic.
func (t *Trie) Lookup(ssid Ssid, filter func(s Subscriber) bool) (subs Subscribers) {
	subs = newSubscribers()
	t.RLock()

	t.lookup(ssid, &subs, t.root, filter)

	if contractNode, ok := t.root.children[ssid[0]]; ok {
		if shareNode, ok := contractNode.children[share]; ok {
			t.randomByGroup(ssid[1:], &subs, shareNode, filter)
		}
	}

	t.RUnlock()
	return
}

func (t *Trie) lookupEmitter(query Ssid, subs *Subscribers, node *node, filter func(s Subscriber) bool) {
	// Add subscribers from the current branch
	subs.AddRange(node.subs, filter)

	// If we're done, stop
	if len(query) == 0 {
		return
	}

	// Go through the exact match branch
	if n, ok := node.children[query[0]]; ok {
		t.lookupEmitter(query[1:], subs, n, filter)
	}

	// Go through wildcard match branch
	if n, ok := node.children[wildcard]; ok {
		t.lookupEmitter(query[1:], subs, n, filter)
	}
}

func (t *Trie) lookupMqtt(query Ssid, subs *Subscribers, node *node, filter func(s Subscriber) bool) {
	// If we're done, stop
	if len(query) == 0 {
		// Add subscribers from the current branch
		subs.AddRange(node.subs, filter)
		return
	}

	// Go through the exact match branch
	if n, ok := node.children[query[0]]; ok {
		t.lookupMqtt(query[1:], subs, n, filter)
	}

	// Go through wildcard match branch
	if n, ok := node.children[wildcard]; ok {
		t.lookupMqtt(query[1:], subs, n, filter)
	}

	// Add subscribers from multi-wildcard branch
	if n, ok := node.children[multiWildcard]; ok {
		subs.AddRange(n.subs, filter)
	}
}

// Reusable pool of subscriber groups
var temp = &sync.Pool{
	New: func() interface{} {
		x := time.Now().UnixNano()
		return &tempState{
			list: newSubscribers(),
			rand: uint32((x >> 32) ^ x),
		}
	},
}

type tempState struct {
	list Subscribers
	rand uint32
}

// RandomByGroup adds a random subscribers for shared subscriptions, by share group
func (t *Trie) randomByGroup(query Ssid, subs *Subscribers, shareNode *node, filter func(s Subscriber) bool) {
	tmp := temp.Get().(*tempState)
	defer temp.Put(tmp)

	// Select a random subscriber from each share group (child of the share node)
	for _, n := range shareNode.children {
		tmp.list.Reset() // recycle
		t.lookup(query, &tmp.list, n, filter)
		if tmp.list.Size() == 0 {
			continue
		}

		// Generate a random number using xorshift
		x := tmp.rand
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		tmp.rand = x

		// Select a random element from the list and add it
		subs.AddUnique(tmp.list.Random(x))

	}
	return
}
