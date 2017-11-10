package mock

import (
	"io"
	"net"
	"time"
)

// End is one 'end' of a simulated connection.
type End struct {
	Reader *io.PipeReader
	Writer *io.PipeWriter
}

// Close closes the End.
func (e End) Close() error {
	if err := e.Writer.Close(); err != nil {
		return err
	}

	return e.Reader.Close()
}

// LocalAddr gets the local address.
func (e End) LocalAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

// RemoteAddr gets the local address.
func (e End) RemoteAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

// Read implements the interface.
func (e End) Read(data []byte) (n int, err error) { return e.Reader.Read(data) }

// Write implements the interface.
func (e End) Write(data []byte) (n int, err error) { return e.Writer.Write(data) }

// SetDeadline implements the interface.
func (e End) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline implements the interface.
func (e End) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline implements the interface.
func (e End) SetWriteDeadline(t time.Time) error { return nil }
