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

import (
	"net"

	"github.com/emitter-io/emitter/logging"
)

var logConnection = logging.AddLogger("[peer] connection %s (remote: %s)")

// Peer represents a peer broker.
type Peer struct {
	socket net.Conn // The transport used to read and write messages.
}

// NewPeer creates a new peer for the connection.
func newPeer(conn net.Conn) *Peer {
	logging.Log(logConnection, "opened", conn.RemoteAddr().String())
	return &Peer{socket: conn}
}

// Send forwards the message to the remote server.
func (c *Peer) Send(channel []byte, payload []byte) error {

	return nil
}

// Close terminates the connection.
func (c *Peer) Close() error {
	logging.Log(logConnection, "closed", c.socket.RemoteAddr().String())
	return c.socket.Close()
}

/*

// NewPeerManager creates a new manager for the peers.
func NewPeerManager(service *Service) *PeerManager {
	return &PeerManager{
		service: service,
		peers:   collection.NewConcurrentMap(),
	}
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
