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
	root *node
	mu   sync.RWMutex
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

// Subscribe adds the Subscriber to the topic and returns a Subscription.
func (t *Trie) Subscribe(ssid Ssid, sub Subscriber) (*Subscription, error) {
	t.mu.Lock()
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

	curr.subs.AddUnique(sub)
	t.mu.Unlock()
	return &Subscription{Ssid: ssid, Subscriber: sub}, nil
}

// Unsubscribe removes the Subscription.
func (t *Trie) Unsubscribe(ssid Ssid, subscriber Subscriber) {
	t.mu.Lock()
	curr := t.root
	for _, word := range ssid {
		child, ok := curr.children[word]
		if !ok {
			// Subscription doesn't exist.
			t.mu.Unlock()
			return
		}
		curr = child
	}
	curr.subs.Remove(subscriber)
	if len(curr.subs) == 0 && len(curr.children) == 0 {
		curr.orphan()
	}
	t.mu.Unlock()
}

// Lookup returns the Subscribers for the given topic.
func (t *Trie) Lookup(query Ssid) Subscribers {
	t.mu.RLock()
	subs := t.lookup(query, t.root)
	t.mu.RUnlock()
	return subs
}

func (t *Trie) lookup(query Ssid, node *node) Subscribers {
	if len(query) == 0 {
		return node.subs
	}

	// Add subscribers from the current branch
	var subs Subscribers
	for _, s := range node.subs {
		subs.AddUnique(s)
	}

	// Go through the exact match branch
	if n, ok := node.children[query[0]]; ok {
		for _, v := range t.lookup(query[1:], n) {
			subs.AddUnique(v)
		}
	}

	// Go through wildcard match branc
	if n, ok := node.children[wildcard]; ok {
		for _, v := range t.lookup(query[1:], n) {
			subs.AddUnique(v)
		}
	}
	return subs
}
