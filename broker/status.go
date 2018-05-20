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
	"fmt"

	"github.com/emitter-io/emitter/monitor"
	"github.com/emitter-io/emitter/network/address"
)

// sendStats writes the stats and publishes it
func (s *Service) sendStats() {
	node := address.Fingerprint(s.LocalName())
	addr := address.External().String()
	stat := s.measurer

	// Track runtime information
	stat.MeasureRuntime()

	// Track node specific information
	stat.Measure("node.id", int64(node))
	stat.Measure("node.peers", int64(s.NumPeers()))
	stat.Measure("node.conns", int64(s.connections))
	//stat.Measure("node.subs", TODO)

	// Add node tags
	stat.Tag("node.id", node.String())
	stat.Tag("node.addr", addr)

	// Create a snaphshot of all stats and publish the stats
	if m, ok := stat.(monitor.Snapshotter); ok {
		s.selfPublish(fmt.Sprintf("stats/%s/", addr), m.Snapshot())
	}
}
