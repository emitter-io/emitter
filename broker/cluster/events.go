package cluster

// SubscriptionEvent represents a message sent when a subscription is added or removed.
type SubscriptionEvent struct {
	Node    string   // Gets or sets the node identifier for this event.
	Ssid    []uint32 // Gets or sets the SSID (parsed channel) for this subscription.
	Channel string   // Gets or sets the channel for this subscription.
}
