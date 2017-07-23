package perf

import (
	"time"
)

// StatusOfEmitter represents the statistics for the emitteio node, published inside emitter itself.
type statusOfEmitter struct {
	Node    string          `json:"node"`
	Host    string          `json:"host"`
	Time    time.Time       `json:"time"`
	Machine string          `json:"machine"`
	CPU     float64         `json:"cpu"`
	Network statusOfNetwork `json:"network"`
}

// StatusOfNetwork represents the statistics for the emitter node, published inside emitter itself.
type statusOfNetwork struct {
	Connections        int     `json:"connections"`
	AveragePPSIncoming float64 `json:"avg-pps-in"`
	AveragePPSOutgoing float64 `json:"avg-pps-out"`
	AverageMPSIncoming float64 `json:"avg-mps-in"`
	AverageMPSOutgoing float64 `json:"avg-mps-out"`
	AverageBPSIncoming float64 `json:"avg-bps-in"`
	AverageBPSOutgoing float64 `json:"avg-bps-out"`
	Compression        float64 `json:"compression"`
}

// NewNetworkCounters creates a shortcut structure for network counters.
func (c *Counters) NewNetworkCounters() *NetworkCounters {
	return &NetworkCounters{
		MessagesIn:  c.GetCounter("net.message.in"),
		MessagesOut: c.GetCounter("net.message.out"),
		PacketsIn:   c.GetCounter("net.packet.in"),
		PacketsOut:  c.GetCounter("net.packet.out"),
		TrafficIn:   c.GetCounter("net.traffic.in"),
		TrafficOut:  c.GetCounter("net.traffic.out"),
	}
}

// NetworkCounters represents a shortcut structure for some of the frequently used network counters.
type NetworkCounters struct {
	MessagesIn  *Counter // The counter for incoming messages.
	MessagesOut *Counter // The counter for outgoing messages.
	PacketsIn   *Counter // The counter for incoming packets.
	PacketsOut  *Counter // The counter for outgoing packets.
	TrafficIn   *Counter // The counter for incoming traffic.
	TrafficOut  *Counter // The counter for outgoing traffic.
}
