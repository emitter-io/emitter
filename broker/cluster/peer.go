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

package cluster

/*
import (
	"net"

	"github.com/emitter-io/emitter/collection"
	"github.com/emitter-io/emitter/utils"
	"github.com/hashicorp/serf/serf"
)

// Peer represents a peer broker.
type Peer struct {
	service *Service // The service for this connection.
	socket  net.Conn // The transport used to read and write messages.
}

// PeerManager manages the emitter broker peers
type PeerManager struct {
	service *Service                  // The service for this connection.
	peers   *collection.ConcurrentMap // The internal map of the peers.
}

// NewPeerManager creates a new manager for the peers.
func NewPeerManager(service *Service) *PeerManager {
	return &PeerManager{
		service: service,
		peers:   collection.NewConcurrentMap(),
	}
}

// NewPeer creates a new peer for the connection.
func (m *PeerManager) newPeer(conn net.Conn) *Peer {
	return &Peer{service: m.service, socket: conn}
}

// GetMember retrieves the member by its id
func (m *PeerManager) getMember(node string) *serf.Member {
	cluster := m.service.cluster
	for _, m := range cluster.Members() {
		if m.Name == node {
			return &m
		}
	}

	return nil
}

// Get retrieves a peer from the manager, if not find it attempts to connect to a peer.
func (m *PeerManager) Get(node string) (peer *Peer, ok bool) {
	// First we need to find the member
	member := m.getMember(node)
	if member == nil {
		return nil, false
	}

	// Attempt to retrieve the associated peer
	key := utils.GetHash([]byte(member.Name))
	peer, ok = m.peers.GetOrCreate(key, func() interface{} {
		addr := member.Tags["route"]
		if conn, err := net.Dial("tcp", addr); err != nil {
			return m.newPeer(conn)
		}
		return nil
	}).(*Peer)
	return
}
*/
