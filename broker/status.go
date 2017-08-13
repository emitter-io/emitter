package broker

import (
	"encoding/json"
	"github.com/emitter-io/emitter/logging"
	"time"

	"github.com/emitter-io/emitter/network/address"
	"github.com/kelindar/process"
)

// StatusInfo represents the status payload.
type StatusInfo struct {
	Node          string    `json:"node"`
	Addr          string    `json:"addr"`
	Subscriptions int       `json:"subs"`
	CPU           float64   `json:"cpu"`
	MemoryPrivate int64     `json:"priv"`
	MemoryVirtual int64     `json:"virt"`
	Time          time.Time `json:"time"`
	Uptime        float64   `json:"uptime"`
}

// getStatus retrieves the status of the service.
func (s *Service) getStatus() *StatusInfo {
	stats := new(StatusInfo)

	// Fill the identity
	t := time.Now().UTC()
	stats.Node = s.LocalName()
	stats.Addr = address.External().String()
	stats.Subscriptions = s.subcounters.Count()
	stats.Time = t
	stats.Uptime = t.Sub(s.startTime).Seconds()

	// Collect CPU and Memory stats
	process.ProcUsage(&stats.CPU, &stats.MemoryPrivate, &stats.MemoryVirtual)
	return stats
}

// Reports the status periodically.
func (s *Service) reportStatus() {
	status := s.getStatus()
	b, err := json.Marshal(status)
	if err != nil {
		logging.LogError("service", "reporting status", err)
	}

	s.selfPublish("cluster/"+status.Addr+"/", b)
}
