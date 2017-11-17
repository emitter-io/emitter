package mock

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnd(t *testing.T) {
	r, w := io.Pipe()
	end := End{
		Reader: r,
		Writer: w,
	}

	assert.Equal(t, "127.0.0.1", end.LocalAddr().String())
	assert.Equal(t, "127.0.0.1", end.RemoteAddr().String())
	assert.NoError(t, end.Close())
	assert.NoError(t, end.SetDeadline(time.Now()))
	assert.NoError(t, end.SetReadDeadline(time.Now()))
	assert.NoError(t, end.SetWriteDeadline(time.Now()))

}
