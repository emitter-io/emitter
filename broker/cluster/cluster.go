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
	"errors"
	"fmt"
	"net"

	"github.com/emitter-io/emitter/collection"
	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/encoding"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/emitter/network/tcp"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
)

// Cluster represents a cluster manager.
type Cluster struct {
	name            string                         // The name of the local node.
	closing         chan bool                      // The closing channel.
	gossip          *serf.Serf                     // The gossip-based cluster mechanism.
	config          *serf.Config                   // The configuration for gossip.
	peers           *collection.ConcurrentMap      // The internal map of the peers.
	events          chan serf.Event                // The channel for receiving gossip events.
	Subscriptions   func() []Subscription          // Delegate to retrieve all existing subscriptions.
	OnQuery         func(Query)                    // Delegate to invoke when a query is received.
	OnQueryResponse func(QueryResponse)            // Delegate to invoke when a query is received.
	OnSubscribe     func(*Peer, SubscriptionEvent) // Delegate to invoke when the subscription event is received.
	OnUnsubscribe   func(*Peer, SubscriptionEvent) // Delegate to invoke when the subscription event is received.
	OnMessage       func(*Message)                 // Delegate to invoke when a new message is received.
}

// NewCluster creates a new cluster manager.
func NewCluster(cfg *config.ClusterConfig, closing chan bool) (*Cluster, error) {
	cluster := new(Cluster)
	cluster.events = make(chan serf.Event)
	cluster.closing = closing
	cluster.peers = collection.NewConcurrentMap()
	if err := cluster.configure(cfg); err != nil {
		return nil, err
	}

	return cluster, nil
}

// Listen creates the listener and serves the cluster.
func (c *Cluster) Listen(port int) (err error) {
	if c.gossip, err = serf.Create(c.config); err != nil {
		return
	}

	// Listen on cluster event loop
	go c.clusterEventLoop()

	// Start serving
	err = tcp.ServeAsync(port, c.closing, c.onAccept)
	return
}

// LocalName returns the local node name.
func (c *Cluster) LocalName() string {
	return c.name
}

/*
func (c *Cluster) querySubs() {
	msg, err := encoding.Encode(&QueryEvent{Node: c.LocalName()})
	if err != nil {
		logging.LogError("cluster", "query encoding", err)
	}

	resp, err := c.gossip.Query("subs", msg, nil)
	if err != nil {
		logging.LogError("cluster", "query", err)
	}

	// Wait for all the responses to come back to us before starting.
	for r := range resp.ResponseCh() {
		c.OnQueryResponse(QueryResponse{
			Node:    r.From,
			Payload: r.Payload,
		})
	}
}*/

// Creates a configuration for the cluster
func (c *Cluster) configure(cfg *config.ClusterConfig) error {
	config := serf.DefaultConfig()
	config.RejoinAfterLeave = true
	config.NodeName = address.Fingerprint() //fmt.Sprintf("%s:%d", address.External().String(), cfg.Cluster.Port) // TODO: fix this
	config.EventCh = c.events
	config.SnapshotPath = cfg.SnapshotPath
	config.MemberlistConfig = memberlist.DefaultWANConfig()
	config.MemberlistConfig.BindPort = cfg.Gossip
	config.MemberlistConfig.AdvertisePort = cfg.Gossip
	config.MemberlistConfig.SecretKey = cfg.Key()

	// Set the node name
	config.NodeName = cfg.NodeName
	if config.NodeName == "" {
		config.NodeName = fmt.Sprintf("%s%d", address.Fingerprint(), cfg.Gossip)
	}
	c.name = config.NodeName

	// Use the public IP address if necessary
	if cfg.AdvertiseAddr == "public" {
		config.MemberlistConfig.AdvertiseAddr = address.External().String()
	}

	// Configure routing
	config.Tags = make(map[string]string)
	config.Tags["route"] = fmt.Sprintf("%s:%d", config.MemberlistConfig.AdvertiseAddr, cfg.Route)
	c.config = config
	return nil
}

// Listens to incoming cluster events.
func (c *Cluster) clusterEventLoop() {
	for {
		select {
		case <-c.closing:
			return
		case e := <-c.events:
			switch e.EventType() {

			// Handles when a new member joins the cluster. When this happens, we need to
			// try to connect to the new node for message forwarding.
			case serf.EventMemberJoin:
				event := e.(serf.MemberEvent)
				for _, m := range event.Members {
					c.peerConnect(m)
				}

			case serf.EventQuery:
				if event, ok := e.(*serf.Query); ok {
					c.onQueryEvent(event)
				}

			// Handles user event which in this case is subscription or unsubscription.
			case serf.EventUser:
				event := e.(serf.UserEvent)
				if err := c.onUserEvent(&event); err != nil {
					logging.LogError("cluster", "event received", err)
				}
			}
		}
	}
}

// Join attempts to join a set of existing peers.
func (c *Cluster) Join(peers ...string) error {
	_, err := c.gossip.Join(peers, true)
	return err
}

// Broadcast is used to broadcast a custom user event with a given name and
// payload. The events must be fairly small, and if the  size limit is exceeded
// and error will be returned. If coalesce is enabled, nodes are allowed to
// coalesce this event.
func (c *Cluster) Broadcast(name string, message interface{}) error {
	buffer, err := encoding.Encode(message)
	if err != nil {
		return err
	}

	return c.gossip.UserEvent(name, buffer, true)
}

// Occurs when a new cluster event is received.
func (c *Cluster) onQueryEvent(e *serf.Query) {
	c.OnQuery(Query{
		Name:    e.Name,
		Respond: e.Respond,
	})
}

// Occurs when a new cluster event is received.
func (c *Cluster) onUserEvent(e *serf.UserEvent) error {
	switch e.Name {
	case "+":
		// This is a subscription event which occurs when a client is subscribed to a node.
		event := *decodeSubscriptionEvent(e.Payload)
		if c.OnSubscribe != nil && event.Node != c.LocalName() {
			if peer, ok := c.getPeer(event.Node); ok {
				c.OnSubscribe(peer, event)
			}
		}

	case "-":
		// This is an unsubscription event which occurs when a client is unsubscribed from a node.
		event := *decodeSubscriptionEvent(e.Payload)
		if c.OnUnsubscribe != nil && event.Node != c.LocalName() {
			if peer, ok := c.getPeer(event.Node); ok {
				c.OnUnsubscribe(peer, event)
			}
		}

	default:
		return errors.New("received unknown event name: " + e.Name)
	}

	return nil
}

// Occurs when a new handshake is received. This allows us to validate the handshake and
// at least check whether a peer with this node name is already connected or not.
func (c *Cluster) onHandshake(peer *Peer, e HandshakeEvent) error {
	if _, exists := c.peers.Get(peerKey(e.Node)); exists {
		return errors.New("Already connected to peer " + e.Node)
	}

	// Accepted the handshake, add to our registry
	logging.LogAction("cluster", "handshake accepted from "+e.Node)
	peer.name = e.Node
	c.peers.Set(peerKey(e.Node), peer)

	for _, s := range e.Subs {
		println("handshake subscription: " + string(s.Channel))
		c.OnSubscribe(peer, SubscriptionEvent{
			Ssid:    s.Ssid,
			Channel: string(s.Channel),
			Node:    e.Node,
		})
	}

	// Handshake with the peer (will only be sent once)
	peer.Handshake(c.LocalName(), c.Subscriptions)
	return nil
}

// Occurs when a new peer connection is accepted.
func (c *Cluster) onAccept(conn net.Conn) {
	peer := newPeer(conn)
	peer.OnHandshake = c.onHandshake
	peer.OnMessage = c.OnMessage
	peer.OnClosing = c.onPeerClosing
	go peer.Process()
}

// PeerConnect connects to the peer node.
func (c *Cluster) peerConnect(node serf.Member) {
	addr := node.Tags["route"]
	if c.gossip.LocalMember().Tags["route"] == addr {
		return
	}

	// Dial the peer who just joined
	logging.LogAction("cluster", "connecting to "+addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logging.LogError("cluster", "peer connect", err)
		return
	}

	// Create a new peer with the appropriate delegates attached.
	peer := newPeer(conn)
	peer.OnHandshake = c.onHandshake
	peer.OnMessage = c.OnMessage
	peer.OnClosing = c.onPeerClosing
	go peer.Process()

	// Send the handshake through
	peer.Handshake(c.LocalName(), c.Subscriptions) // TODO check error
}

// Occurs when a peer is closing
func (c *Cluster) onPeerClosing(p *Peer) {
	println("peer closing: " + p.name)
	if p.name == "" {
		return
	}

	if peer, exists := c.getPeer(p.name); exists && p == peer {
		println("peer " + p.name + " deleted")
		c.peers.Delete(peerKey(p.name))
	}
}

// GetPeer retrieves a peer from the registry.
func (c *Cluster) getPeer(name string) (*Peer, bool) {
	if v, ok := c.peers.Get(peerKey(name)); ok {
		return v.(*Peer), true
	}

	return nil, false
}

// Close terminates/leaves the cluster gracefully.
func (c *Cluster) Close() (err error) {
	if c.gossip != nil {
		err = c.gossip.Leave()
		err = c.gossip.Shutdown()
	}
	return
}
