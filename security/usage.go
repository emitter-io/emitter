package security

import (
	"sync/atomic"
)

// UsageStats represents a tracker for incoming and outgoing traffic.
type UsageStats interface {
	AddIngress(size int64)      // Records the ingress message size.
	AddEgress(size int64)       // Records the egress message size.
	GetIngress() (int64, int64) // Returns the number of ingress messages and bytes recorded.
	GetEgress() (int64, int64)  // Returns the number of egress messages and bytes recorded.
	Reset()                     // Resets the tracker.
}

// NewUsageStats constructs a new usage statistics instance.
func NewUsageStats() UsageStats {
	return new(usage)
}

type usage struct {
	messageIn int64
	trafficIn int64
	messageEg int64
	trafficEg int64
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

// Resets the tracker.
func (t *usage) Reset() {
	atomic.StoreInt64(&t.messageIn, 0)
	atomic.StoreInt64(&t.trafficIn, 0)
	atomic.StoreInt64(&t.messageEg, 0)
	atomic.StoreInt64(&t.trafficEg, 0)
}
