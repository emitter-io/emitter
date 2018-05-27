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
	"io"

	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/stats"
)

// sampler reads statistics of the service and creates a snapshot
type sampler struct {
	service  *Service       // The service to use for stats collection.
	measurer stats.Measurer // The measurer to use for snapshotting.
}

// newSampler creates a stats sampler.
func newSampler(s *Service, m stats.Measurer) io.Reader {
	return &sampler{
		service:  s,
		measurer: m,
	}
}

// Read reads the stats snapshot and writes it to the output buffer.
func (s *sampler) Read(p []byte) (n int, err error) {
	stat := s.service.measurer
	serv := s.service
	node := address.Fingerprint(serv.LocalName())
	addr := address.External().String()

	// Track runtime information
	stat.MeasureRuntime()

	// Track node specific information
	stat.Measure("node.id", int32(node))
	stat.Measure("node.peers", int32(serv.NumPeers()))
	stat.Measure("node.conns", int32(serv.connections))
	stat.Measure("node.subs", int32(serv.subscriptions.Count()))

	// Add node tags
	stat.Tag("node.id", node.String())
	stat.Tag("node.addr", addr)

	// Create a snaphshot of all stats and write it
	if m, ok := stat.(stats.Snapshotter); ok {
		snapshot := m.Snapshot()
		if n = len(snapshot); n > 0 {
			copy(p, snapshot)
		}
	}

	err = io.EOF
	return
}
