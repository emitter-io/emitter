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
	"sync/atomic"
	"unsafe"
)

const wildcard = uint32(1815237614)

type iNode struct {
	main *mainNode
}

type mainNode struct {
	cNode *cNode
	tNode *tNode
}

type cNode struct {
	branches map[uint32]*branch
}

// newCNode creates a new C-node with the given subscription path.
func newCNode(words []uint32, sub Subscriber) *cNode {
	if len(words) == 1 {
		return &cNode{
			branches: map[uint32]*branch{
				words[0]: {subs: Subscribers{sub}},
			},
		}
	}

	nin := &iNode{main: &mainNode{cNode: newCNode(words[1:], sub)}}
	return &cNode{
		branches: map[uint32]*branch{
			words[0]: {subs: Subscribers{}, iNode: nin},
		},
	}
}

// inserted returns a copy of this C-node with the specified Subscriber
// inserted.
func (c *cNode) inserted(words []uint32, sub Subscriber) *cNode {
	branches := make(map[uint32]*branch, len(c.branches)+1)
	for key, branch := range c.branches {
		branches[key] = branch
	}
	var br *branch
	if len(words) == 1 {
		br = &branch{subs: Subscribers{sub}}
	} else {
		br = &branch{
			subs:  Subscribers{},
			iNode: &iNode{main: &mainNode{cNode: newCNode(words[1:], sub)}},
		}
	}
	branches[words[0]] = br
	return &cNode{branches: branches}
}

// updated returns a copy of this C-node with the specified branch updated.
func (c *cNode) updated(word uint32, sub Subscriber) *cNode {
	branches := make(map[uint32]*branch, len(c.branches))
	for word, branch := range c.branches {
		branches[word] = branch
	}

	newBranch := &branch{subs: Subscribers{sub}}
	if br, ok := branches[word]; ok {
		newBranch.iNode = br.iNode
		for _, sub := range br.subs {
			newBranch.subs = append(newBranch.subs, sub)
		}
	}

	branches[word] = newBranch
	return &cNode{branches: branches}
}

// updatedBranch returns a copy of this C-node with the specified branch
// updated.
func (c *cNode) updatedBranch(word uint32, in *iNode, br *branch) *cNode {
	branches := make(map[uint32]*branch, len(c.branches))
	for key, branch := range c.branches {
		branches[key] = branch
	}
	branches[word] = br.updated(in)
	return &cNode{branches: branches}
}

// removed returns a copy of this C-node with the Subscriber removed from the
// corresponding branch.
func (c *cNode) removed(word uint32, sub Subscriber) *cNode {
	branches := make(map[uint32]*branch, len(c.branches))
	for word, branch := range c.branches {
		branches[word] = branch
	}
	br, ok := branches[word]
	if ok {
		br = br.removed(sub)
		if len(br.subs) == 0 && br.iNode == nil {
			// Remove the branch if it contains no subscribers and doesn't
			// point anywhere.
			delete(branches, word)
		} else {
			branches[word] = br
		}
	}
	return &cNode{branches: branches}
}

// getBranches returns the branches for the given word. There are two possible
// branches: exact match and single wildcard.
func (c *cNode) getBranches(word uint32) (*branch, *branch) {
	return c.branches[word], c.branches[wildcard]
}

type branch struct {
	iNode *iNode
	subs  Subscribers
}

// updated returns a copy of this branch updated with the given I-node.
func (b *branch) updated(in *iNode) *branch {
	subs := make(Subscribers, len(b.subs))
	copy(subs, b.subs)
	return &branch{subs: subs, iNode: in}
}

// removed returns a copy of this branch with the given Subscriber removed.
func (b *branch) removed(sub Subscriber) *branch {
	subs := make(Subscribers, 0)
	for _, s := range b.subs {
		if s != sub {
			subs = append(subs, s)
		}
	}

	return &branch{subs: subs, iNode: b.iNode}
}

// subscribers returns the Subscribers for this branch.
func (b *branch) subscribers() Subscribers {
	subs := make(Subscribers, len(b.subs))
	copy(subs, b.subs)
	return subs
}

type tNode struct{}

// SubscriptionTrie represents an efficient collection of subscriptions with lookup capability.
type SubscriptionTrie struct {
	root *iNode
}

// NewSubscriptionTrie creates a new matcher for the subscriptions.
func NewSubscriptionTrie() *SubscriptionTrie {
	root := &iNode{main: &mainNode{cNode: &cNode{}}}
	return &SubscriptionTrie{root: root}
}

// Subscribe adds the Subscriber to the topic and returns a Subscription.
func (c *SubscriptionTrie) Subscribe(ssid Ssid, sub Subscriber) (*Subscription, error) {
	rootPtr := (*unsafe.Pointer)(unsafe.Pointer(&c.root))
	root := (*iNode)(atomic.LoadPointer(rootPtr))
	if !c.iinsert(root, nil, ssid, sub) {
		c.Subscribe(ssid, sub)
	}

	return &Subscription{
		Ssid:       ssid,
		Subscriber: sub,
	}, nil
}

func (c *SubscriptionTrie) iinsert(i, parent *iNode, words []uint32, sub Subscriber) bool {
	// Linearization point.
	mainPtr := (*unsafe.Pointer)(unsafe.Pointer(&i.main))
	main := (*mainNode)(atomic.LoadPointer(mainPtr))
	switch {
	case main.cNode != nil:
		cn := main.cNode
		if br := cn.branches[words[0]]; br == nil {
			// If the relevant branch is not in the map, a copy of the C-node
			// with the new entry is created. The linearization point is a
			// successful CAS.
			ncn := &mainNode{cNode: cn.inserted(words, sub)}
			return atomic.CompareAndSwapPointer(
				mainPtr, unsafe.Pointer(main), unsafe.Pointer(ncn))
		} else {
			// If the relevant key is present in the map, its corresponding
			// branch is read.
			if len(words) > 1 {
				// If more than 1 word is present in the path, the tree must be
				// traversed deeper.
				if br.iNode != nil {
					// If the branch has an I-node, iinsert is called
					// recursively.
					return c.iinsert(br.iNode, i, words[1:], sub)
				}
				// Otherwise, an I-node which points to a new C-node must be
				// added. The linearization point is a successful CAS.
				nin := &iNode{main: &mainNode{cNode: newCNode(words[1:], sub)}}
				ncn := &mainNode{cNode: cn.updatedBranch(words[0], nin, br)}
				return atomic.CompareAndSwapPointer(
					mainPtr, unsafe.Pointer(main), unsafe.Pointer(ncn))
			}

			if br.subs.Contains(sub) {
				// Already subscribed.
				return true
			}

			// Insert the Subscriber by copying the C-node and updating the
			// respective branch. The linearization point is a successful CAS.
			ncn := &mainNode{cNode: cn.updated(words[0], sub)}
			return atomic.CompareAndSwapPointer(
				mainPtr, unsafe.Pointer(main), unsafe.Pointer(ncn))
		}
	case main.tNode != nil:
		clean(parent)
		return false
	default:
		panic("SubscriptionTrie is in an invalid state")
	}
}

// Unsubscribe removes the Subscription.
func (c *SubscriptionTrie) Unsubscribe(ssid Ssid, subscriber Subscriber) {
	rootPtr := (*unsafe.Pointer)(unsafe.Pointer(&c.root))
	root := (*iNode)(atomic.LoadPointer(rootPtr))

	if !c.iremove(root, nil, nil, ssid, 0, subscriber) {
		c.Unsubscribe(ssid, subscriber)
	}
}

func (c *SubscriptionTrie) iremove(i, parent, parentsParent *iNode, words []uint32,
	wordIdx int, sub Subscriber) bool {

	// Linearization point.
	mainPtr := (*unsafe.Pointer)(unsafe.Pointer(&i.main))
	main := (*mainNode)(atomic.LoadPointer(mainPtr))
	switch {
	case main.cNode != nil:
		cn := main.cNode
		br := cn.branches[words[wordIdx]]
		if br == nil {
			// If the relevant word is not in the map, the subscription doesn't
			// exist.
			return true
		}

		// If the relevant word is present in the map, its corresponding
		// branch is read.
		if wordIdx+1 < len(words) {
			// If more than 1 word is present in the path, the tree must be
			// traversed deeper.
			if br.iNode != nil {
				// If the branch has an I-node, iremove is called
				// recursively.
				return c.iremove(br.iNode, i, parent, words, wordIdx+1, sub)
			}
			// Otherwise, the subscription doesn't exist.
			return true
		}

		if !br.subs.Contains(sub) {
			// Not subscribed.
			return true
		}

		// Remove the Subscriber by copying the C-node without it. A
		// contraction of the copy is then created. A successful CAS will
		// substitute the old C-node with the copied C-node, thus removing
		// the Subscriber from the trie - this is the linearization point.
		ncn := cn.removed(words[wordIdx], sub)
		cntr := c.toContracted(ncn, i)
		if atomic.CompareAndSwapPointer(
			mainPtr, unsafe.Pointer(main), unsafe.Pointer(cntr)) {
			if parent != nil {
				mainPtr := (*unsafe.Pointer)(unsafe.Pointer(&i.main))
				main := (*mainNode)(atomic.LoadPointer(mainPtr))
				if main.tNode != nil {
					cleanParent(i, parent, parentsParent, c, words[wordIdx-1])
				}
			}
			return true
		}
		return false

	case main.tNode != nil:
		clean(parent)
		return false
	default:
		panic("SubscriptionTrie is in an invalid state")
	}
}

// Lookup returns the Subscribers for the given topic.
func (c *SubscriptionTrie) Lookup(query Ssid) Subscribers {
	rootPtr := (*unsafe.Pointer)(unsafe.Pointer(&c.root))
	root := (*iNode)(atomic.LoadPointer(rootPtr))
	subs := make(Subscribers, 0, 6)

	if ok := c.ilookup(root, nil, query, &subs); ok {
		return subs
	}

	return c.Lookup(query)
}

// ilookup attempts to retrieve the Subscribers for the word path. True is
// returned if the Subscribers were retrieved, false if the operation needs to
// be retried.
func (c *SubscriptionTrie) ilookup(i, parent *iNode, words []uint32, subs *Subscribers) bool {
	// Linearization point.
	mainPtr := (*unsafe.Pointer)(unsafe.Pointer(&i.main))
	main := (*mainNode)(atomic.LoadPointer(mainPtr))

	switch {
	case main.cNode != nil:
		// Traverse exact-match branch and single-word-wildcard branch.
		exact, singleWC := main.cNode.getBranches(words[0])
		if exact != nil {
			if !c.bLookup(i, parent, main, exact, words, subs) {
				return false
			}
		}

		if singleWC != nil {
			if !c.bLookup(i, parent, main, singleWC, words, subs) {
				return false
			}
		}

		return true
	case main.tNode != nil:
		clean(parent)
		return false
	default:
		panic("Subscription Trie is in an invalid state")
	}
}

// bLookup attempts to retrieve the Subscribers from the word path along the
// given branch. True is returned if the Subscribers were retrieved, false if
// the operation needs to be retried.
func (c *SubscriptionTrie) bLookup(i, parent *iNode, main *mainNode, b *branch, words []uint32, subs *Subscribers) bool {
	// Retrieve the subscribers from the branch we are currently traversing.
	if len(b.subs) > 0 {
		for _, s := range b.subscribers() {
			subs.AddUnique(s)
		}
	}

	if len(words) > 1 {
		// If more than 1 key is present in the path, the tree must be
		// traversed deeper.
		if b.iNode == nil {
			// If the branch doesn't point to an I-node, no subscribers exist.
			return true
		}

		// If the branch has an I-node, ilookup is called recursively.
		return c.ilookup(b.iNode, i, words[1:], subs)
	}
	return true
}

// toContracted ensures that every I-node except the root points to a C-node
// with at least one branch or a T-node. If a given C-node has no branches and
// is not at the root level, a T-node is returned.
func (c *SubscriptionTrie) toContracted(cn *cNode, parent *iNode) *mainNode {
	if c.root != parent && len(cn.branches) == 0 {
		return &mainNode{tNode: &tNode{}}
	}
	return &mainNode{cNode: cn}
}

// clean replaces an I-node's C-node with a copy that has any tombed I-nodes
// resurrected.
func clean(i *iNode) {
	mainPtr := (*unsafe.Pointer)(unsafe.Pointer(&i.main))
	main := (*mainNode)(atomic.LoadPointer(mainPtr))
	if main.cNode != nil {
		atomic.CompareAndSwapPointer(mainPtr,
			unsafe.Pointer(main), unsafe.Pointer(toCompressed(main.cNode)))
	}
}

// cleanParent reads the main node of the parent I-node p and the current
// I-node i and checks if the T-node below i is reachable from p. If i is no
// longer reachable, some other thread has already completed the contraction.
// If it is reachable, the C-node below p is replaced with its contraction.
func cleanParent(i, parent, parentsParent *iNode, c *SubscriptionTrie, word uint32) {
	var (
		mainPtr  = (*unsafe.Pointer)(unsafe.Pointer(&i.main))
		main     = (*mainNode)(atomic.LoadPointer(mainPtr))
		pMainPtr = (*unsafe.Pointer)(unsafe.Pointer(&parent.main))
		pMain    = (*mainNode)(atomic.LoadPointer(pMainPtr))
	)
	if pMain.cNode != nil {
		if br, ok := pMain.cNode.branches[word]; ok {
			if br.iNode != i {
				return
			}
			if main.tNode != nil {
				if !contract(parentsParent, parent, i, c, pMain) {
					cleanParent(parentsParent, parent, i, c, word)
				}
			}
		}
	}
}

// contract performs a contraction of the parent's C-node if possible. Returns
// true if the contraction succeeded, false if it needs to be retried.
func contract(parentsParent, parent, i *iNode, c *SubscriptionTrie, pMain *mainNode) bool {
	ncn := toCompressed(pMain.cNode)
	if len(ncn.cNode.branches) == 0 && parentsParent != nil {
		// If the compressed C-node has no branches, it and the I-node above it
		// should be removed. To do this, a CAS must occur on the parent I-node
		// of the parent to update the respective branch of the C-node below it
		// to point to nil.
		ppMainPtr := (*unsafe.Pointer)(unsafe.Pointer(&parentsParent.main))
		ppMain := (*mainNode)(atomic.LoadPointer(ppMainPtr))
		for pKey, pBranch := range ppMain.cNode.branches {
			// Find the branch pointing to the parent.
			if pBranch.iNode == parent {
				// Update the branch to point to nil.
				updated := ppMain.cNode.updatedBranch(pKey, nil, pBranch)
				if len(pBranch.subs) == 0 {
					// If the branch has no subscribers, simply prune it.
					delete(updated.branches, pKey)
				}
				// Replace the main node of the parent's parent.
				return atomic.CompareAndSwapPointer(ppMainPtr,
					unsafe.Pointer(ppMain), unsafe.Pointer(toCompressed(updated)))
			}
		}
	} else {
		// Otherwise, perform a simple contraction to a T-node.
		cntr := c.toContracted(ncn.cNode, parent)
		pMainPtr := (*unsafe.Pointer)(unsafe.Pointer(&parent.main))
		pMain := (*mainNode)(atomic.LoadPointer(pMainPtr))
		if !atomic.CompareAndSwapPointer(pMainPtr, unsafe.Pointer(pMain),
			unsafe.Pointer(cntr)) {
			return false
		}
	}
	return true
}

// toCompressed prunes any branches to tombed I-nodes and returns the
// compressed main node.
func toCompressed(cn *cNode) *mainNode {
	branches := make(map[uint32]*branch, len(cn.branches))
	for key, br := range cn.branches {
		if !prunable(br) {
			branches[key] = br
		}
	}
	return &mainNode{cNode: &cNode{branches: branches}}
}

// prunable indicates if the branch can be pruned. A branch can be pruned if
// it has no subscribers and points to nowhere or it has no subscribers and
// points to a tombed I-node.
func prunable(br *branch) bool {
	if len(br.subs) > 0 {
		return false
	}
	if br.iNode == nil {
		return true
	}
	mainPtr := (*unsafe.Pointer)(unsafe.Pointer(&br.iNode.main))
	main := (*mainNode)(atomic.LoadPointer(mainPtr))
	return main.tNode != nil
}
