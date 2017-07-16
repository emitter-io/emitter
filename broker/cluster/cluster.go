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
	name          string                   // The name of the local node.
	closing       chan bool                // The closing channel.
	gossip        *serf.Serf               // The gossip-based cluster mechanism.
	config        *serf.Config             // The configuration for gossip.
	events        chan serf.Event          // The channel for receiving gossip events.
	OnSubscribe   func(*SubscriptionEvent) // Delegate to invoke when the subscription event is received.
	OnUnsubscribe func(*SubscriptionEvent) // Delegate to invoke when the subscription event is received.
}

// NewCluster creates a new cluster manager.
func NewCluster(cfg *config.ClusterConfig, closing chan bool) (*Cluster, error) {
	cluster := new(Cluster)
	cluster.events = make(chan serf.Event)
	cluster.closing = closing
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
	err = tcp.ServeAsync(port, c.closing, c.onAcceptPeer)
	return
}

// LocalName returns the local node name.
func (c *Cluster) LocalName() string {
	return c.name
}

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
			if e.EventType() == serf.EventUser {
				event := e.(serf.UserEvent)
				if err := c.onEvent(&event); err != nil {
					logging.LogError("service", "event received", err)
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
func (c *Cluster) onEvent(e *serf.UserEvent) error {
	switch e.Name {
	case "+":
		// This is a subscription event which occurs when a client is subscribed to a node.
		var event SubscriptionEvent
		encoding.Decode(e.Payload, &event)

		if event.Node != c.LocalName() {
			c.OnSubscribe(&event)
		}

	case "-":
		// This is an unsubscription event which occurs when a client is unsubscribed from a node.
		var event SubscriptionEvent
		encoding.Decode(e.Payload, &event)

		if event.Node != c.LocalName() {
			c.OnUnsubscribe(&event)
		}

	default:
		return errors.New("received unknown event name: " + e.Name)
	}

	return nil
}

// Occurs when a new peer connection is accepted.
func (c *Cluster) onAcceptPeer(t net.Conn) {

}

// Close terminates/leaves the cluster gracefully.
func (c *Cluster) Close() (err error) {
	if c.gossip != nil {
		err = c.gossip.Leave()
		err = c.gossip.Shutdown()
	}
	return
}
