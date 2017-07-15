package broker

import (
	"net"
)

// Peer represents a peer broker.
type Peer struct {
	socket  net.Conn // The transport used to read and write messages.
	service *Service // The service for this connection.
}
