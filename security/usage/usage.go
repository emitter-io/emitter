package usage

import (
	"sync/atomic"
)

// Meter represents a tracker for incoming and outgoing traffic.
type Meter interface {
	GetContract() uint32   // Returns the associated contract.
	AddIngress(size int64) // Records the ingress message size.
	AddEgress(size int64)  // Records the egress message size.
}

// NewMeter constructs a new usage statistics instance.
func NewMeter(contract uint32) Meter {
	return &usage{Contract: contract}
}

type usage struct {
	MessageIn int64  `json:"mIn"`
	TrafficIn int64  `json:"tIn"`
	MessageEg int64  `json:"mEg"`
	TrafficEg int64  `json:"tEg"`
	Contract  uint32 `json:"id"`
}

// GetContract returns the associated contract.
func (t *usage) GetContract() uint32 {
	return t.Contract
}

// Records the ingress message size.
func (t *usage) AddIngress(size int64) {
	atomic.AddInt64(&t.MessageIn, 1)
	atomic.AddInt64(&t.TrafficIn, size)
}

// Records the egress message size.
func (t *usage) AddEgress(size int64) {
	atomic.AddInt64(&t.MessageEg, 1)
	atomic.AddInt64(&t.TrafficEg, size)
}

// reset resets the tracker and returns old usage.
func (t *usage) reset() *usage {
	var old usage
	old.Contract = t.Contract
	old.MessageIn = atomic.SwapInt64(&t.MessageIn, 0)
	old.TrafficIn = atomic.SwapInt64(&t.TrafficIn, 0)
	old.MessageEg = atomic.SwapInt64(&t.MessageEg, 0)
	old.TrafficEg = atomic.SwapInt64(&t.TrafficEg, 0)
	return &old
}
