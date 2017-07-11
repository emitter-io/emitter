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

// StatusReporter represents a status reporter which periodically reports statistics
type StatusReporter struct {
}
