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
	"encoding/json"
	"fmt"
	"time"

	"github.com/emitter-io/emitter/network/address"
	"github.com/kelindar/process"
)

// StatusInfo represents the status payload.
type StatusInfo struct {
	Node          string    `json:"node"`
	Addr          string    `json:"addr"`
	Subscriptions int       `json:"subs"`
	Connections   int64     `json:"conns"`
	CPU           float64   `json:"cpu"`
	MemoryPrivate int64     `json:"priv"`
	MemoryVirtual int64     `json:"virt"`
	Time          time.Time `json:"time"`
	NumPeers      int       `json:"peers"`
	Uptime        float64   `json:"uptime"`
}

// getStatus retrieves the status of the service.
func (s *Service) getStatus() (*StatusInfo, error) {
	stats := new(StatusInfo)

	// Fill the identity
	t := time.Now().UTC()
	stats.Node = address.Fingerprint(s.LocalName()).String()
	stats.Addr = address.External().String()
	stats.Subscriptions = 0 // TODO: Set subscriptions
	stats.Connections = s.connections
	stats.Time = t
	stats.Uptime = t.Sub(s.startTime).Seconds()
	stats.NumPeers = s.NumPeers()

	// Collect CPU and Memory stats
	return stats, process.ProcUsage(&stats.CPU, &stats.MemoryPrivate, &stats.MemoryVirtual)
}

// Reports the status periodically.
func (s *Service) reportStatus() {
	if status, err := s.getStatus(); err == nil {
		if b, err := json.Marshal(status); err == nil {
			s.selfPublish("cluster/"+status.Addr+"/", b)
		}
	}
}

// statsWriter represents a writer of stats for a particular service.
type statsWriter struct {
	service *Service
}

// Write writes the stats and publishes it
func (w *statsWriter) Write(snapshot []byte) (int, error) {

	w.service.reportStatus()

	m := w.service.measurer
	m.MeasureValue("node.peers", int64(w.service.NumPeers()))
	m.MeasureValue("node.conns", int64(w.service.connections))

	// Publish to the stats channel
	w.service.selfPublish(
		fmt.Sprintf("stats/%s/", address.External().String()),
		snapshot)
	return len(snapshot), nil
}
