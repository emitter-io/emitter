package usage

import (
	"sync/atomic"
)

// Meter represents a tracker for incoming and outgoing traffic.
type Meter interface {
	GetContract() uint32        // Returns the associated contract.
	AddIngress(size int64)      // Records the ingress message size.
	AddEgress(size int64)       // Records the egress message size.
	GetIngress() (int64, int64) // Returns the number of ingress messages and bytes recorded.
	GetEgress() (int64, int64)  // Returns the number of egress messages and bytes recorded.
	Reset() Meter               // Resets the tracker.
}

// NewMeter constructs a new usage statistics instance.
func NewMeter(contract uint32) Meter {
	return &usage{contract: contract}
}

type usage struct {
	contract  uint32
	messageIn int64
	trafficIn int64
	messageEg int64
	trafficEg int64
}

// GetContract returns the associated contract.
func (t *usage) GetContract() uint32 {
	return t.contract
}

// Records the ingress message size.
func (t *usage) AddIngress(size int64) {
	atomic.AddInt64(&t.messageIn, 1)
	atomic.AddInt64(&t.trafficIn, size)
}

// Records the egress message size.
func (t *usage) AddEgress(size int64) {
	atomic.AddInt64(&t.messageEg, 1)
	atomic.AddInt64(&t.trafficEg, size)
}

// Returns the number of ingress messages and bytes recorded.
func (t *usage) GetIngress() (int64, int64) {
	return atomic.LoadInt64(&t.messageIn), atomic.LoadInt64(&t.trafficIn)
}

// Returns the number of egress messages and bytes recorded.
func (t *usage) GetEgress() (int64, int64) {
	return atomic.LoadInt64(&t.messageEg), atomic.LoadInt64(&t.trafficEg)
}

// Resets the tracker and returns old usage.
func (t *usage) Reset() Meter {
	var old usage
	old.contract = t.contract
	old.messageIn = atomic.SwapInt64(&t.messageIn, 0)
	old.trafficIn = atomic.SwapInt64(&t.trafficIn, 0)
	old.messageEg = atomic.SwapInt64(&t.messageEg, 0)
	old.trafficEg = atomic.SwapInt64(&t.trafficEg, 0)
	return &old
}
