package mock

// Addr is a fake network interface which implements the net.Addr interface
type Addr struct {
	NetworkString string
	AddrString    string
}

// Network gets the network string.
func (a Addr) Network() string {
	return a.NetworkString
}

// Network gets the addr string.
func (a Addr) String() string {
	return a.AddrString
}
