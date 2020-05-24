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

package broker

import (
	"sync/atomic"

	"github.com/emitter-io/address"
	"github.com/emitter-io/stats"
)

// sampler reads statistics of the service and creates a snapshot
type sampler struct {
	service  *Service       // The service to use for stats collection.
	measurer stats.Measurer // The measurer to use for snapshotting.
}

// newSampler creates a stats sampler.
func newSampler(s *Service, m stats.Measurer) stats.Snapshotter {
	return &sampler{
		service:  s,
		measurer: m,
	}
}

// Snapshot creates the stats snapshot.
func (s *sampler) Snapshot() (snapshot []byte) {
	stat := s.service.measurer
	serv := s.service
	node := address.Fingerprint(serv.ID())
	addr := serv.Config.Addr()

	// Track runtime information
	stat.MeasureRuntime()

	// Track node specific information
	stat.Measure("node.id", int32(node))
	stat.Measure("node.peers", int32(serv.NumPeers()))
	stat.Measure("node.conns", int32(atomic.LoadInt64(&serv.connections)))
	stat.Measure("node.subs", int32(serv.subscriptions.Count()))

	// Add node tags
	stat.Tag("node.id", node.String())
	stat.Tag("node.addr", addr.String())

	// Create a snaphshot of all stats
	if m, ok := stat.(stats.Snapshotter); ok {
		snapshot = m.Snapshot()
	}
	return
}
