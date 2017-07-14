package perf

// NewUsageTracker creates a structure for monitoring usage
func (c *Counters) NewUsageTracker(prefix string) *UsageTracker {
	return &UsageTracker{
		MessagesIn:  c.GetCounter(prefix + ".message.in"),
		MessagesOut: c.GetCounter(prefix + ".message.out"),
		TrafficIn:   c.GetCounter(prefix + ".traffic.in"),
		TrafficOut:  c.GetCounter(prefix + ".traffic.out"),
	}
}

// UsageTracker represents a shortcut structure for monitoring usage of a contract.
type UsageTracker struct {
	MessagesIn  Counter // The counter for incoming messages.
	MessagesOut Counter // The counter for outgoing messages.
	TrafficIn   Counter // The counter for incoming traffic.
	TrafficOut  Counter // The counter for outgoing traffic.
}

// Reset resets all the counters in the usage tracker.
func (u *UsageTracker) Reset() {
	u.MessagesIn.Reset()
	u.MessagesOut.Reset()
	u.TrafficIn.Reset()
	u.TrafficOut.Reset()
}
