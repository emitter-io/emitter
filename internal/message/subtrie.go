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

package message

import (
	"sync"

	"github.com/kelindar/rand"
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
	if len(n.parent.subs) == 0 && len(n.parent.children) == 0 {
		n.parent.orphan()
	}
}

// Trie represents an efficient collection of subscriptions with lookup capability.
type Trie struct {
	sync.RWMutex
	root  *node // The root node of the tree.
	count int   // Number of subscriptions in the trie.
}

// NewTrie creates a new matcher for the subscriptions.
func NewTrie() *Trie {
	return &Trie{
		root: &node{
			subs:     Subscribers{},
			children: make(map[uint32]*node),
		},
	}
}

// Count returns the number of subscriptions.
func (t *Trie) Count() int {
	t.RLock()
	defer t.RUnlock()
	return t.count
}

// Subscribe adds the Subscriber to the topic and returns a Subscription.
func (t *Trie) Subscribe(ssid Ssid, sub Subscriber) (*Subscription, error) {
	t.Lock()
	curr := t.root
	for _, word := range ssid {
		child, ok := curr.children[word]
		if !ok {
			child = &node{
				word:     word,
				subs:     Subscribers{},
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
	return &Subscription{Ssid: ssid, Subscriber: sub}, nil
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
	if len(curr.subs) == 0 && len(curr.children) == 0 {
		curr.orphan()
	}
	t.Unlock()
}

// Lookup returns the Subscribers for the given topic.
func (t *Trie) Lookup(query Ssid, filter func(s Subscriber) bool) (subs Subscribers) {
	t.RLock()
	t.lookup(query, &subs, t.root, filter)
	t.RUnlock()
	return
}

// Random picks a random subscriber.
func (t *Trie) Random(query Ssid, filter func(s Subscriber) bool) (subs Subscribers) {
	t.RLock()
	t.lookup(query, &subs, t.root, filter)
	t.RUnlock()
	if len(subs) > 0 {
		x := rand.Uint32n(uint32(len(subs)))
		subs = subs[x : x+1]
	}
	return
}

func (t *Trie) lookup(query Ssid, subs *Subscribers, node *node, filter func(s Subscriber) bool) {

	// Add subscribers from the current branch
	for _, s := range node.subs {
		if filter == nil || filter(s) {
			subs.AddUnique(s)
		}
	}

	// If we're not yet done, continue
	if len(query) > 0 {

		// Go through the exact match branch
		if n, ok := node.children[query[0]]; ok {
			t.lookup(query[1:], subs, n, filter)
		}

		// Go through wildcard match branc
		if n, ok := node.children[wildcard]; ok {
			t.lookup(query[1:], subs, n, filter)
		}
	}
}
