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

package cluster

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/weaveworks/mesh"
)

// memberlist represents a peer cache
type memberlist struct {
	list sync.Map
	ctor func(mesh.PeerName) *Peer
}

// newMemberlist creates a new memberlist
func newMemberlist(ctor func(mesh.PeerName) *Peer) *memberlist {
	return &memberlist{
		ctor: ctor,
	}
}

// GetOrAdd gets or adds a peer, returns a peer and whether a new peer was added or not
func (m *memberlist) GetOrAdd(name mesh.PeerName) (*Peer, bool) {
	if p, ok := m.list.Load(name); ok {
		return p.(*Peer), false
	}

	// Create new peer and store it
	peer := m.ctor(name)
	v, loaded := m.list.LoadOrStore(name, peer)
	return v.(*Peer), !loaded
}

// Fallback gets a fallback peer for a given peer.
func (m *memberlist) Fallback(name mesh.PeerName) (*Peer, bool) {
	peers := make([]*Peer, 0, 8)
	m.list.Range(func(k, v interface{}) bool {
		if peer := v.(*Peer); peer.IsActive() && peer.name != name {
			peers = append(peers, v.(*Peer))
		}
		return true
	})

	sort.Slice(peers, func(i, j int) bool { return name-peers[i].name > name-peers[j].name })
	if len(peers) > 0 {
		return peers[0], true
	}
	return nil, false
}

// Touch updates the last activity time
func (m *memberlist) Touch(name mesh.PeerName) {
	peer, _ := m.GetOrAdd(name)
	atomic.StoreInt64(&peer.activity, time.Now().Unix())
}

// Contains checks if a peer is in the memberlist
func (m *memberlist) Contains(name mesh.PeerName) bool {
	_, ok := m.list.Load(name)
	return ok
}

// Remove removes the peer from the memberlist
func (m *memberlist) Remove(name mesh.PeerName) (*Peer, bool) {
	if p, ok := m.list.Load(name); ok {
		peer := p.(*Peer)
		m.list.Delete(peer.name)
		atomic.StoreInt64(&peer.activity, 0)
		return peer, true
	}

	return nil, false
}
