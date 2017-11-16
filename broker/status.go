package broker

import (
	"encoding/json"
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
