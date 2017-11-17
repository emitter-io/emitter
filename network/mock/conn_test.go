package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConn(t *testing.T) {
	conn := NewConn()

	buffer := make([]byte, 10)
	go func() {
		conn.Server.Read(buffer)
	}()

	n, err := conn.Client.Write([]byte{1})
	assert.Equal(t, 1, n)
	assert.NoError(t, err)
	assert.NoError(t, conn.Close())

}
