package tcp

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockListener struct{}

func (l *mockListener) Accept() (net.Conn, error) {
	return nil, errors.New("xxx")
}

func (l *mockListener) Close() error {
	return nil
}

func (l *mockListener) Addr() net.Addr {
	return nil
}

func TestServe(t *testing.T) {
	s := new(Server)
	defer s.Close()
	err := s.Serve(new(mockListener))

	assert.Error(t, err)
}
