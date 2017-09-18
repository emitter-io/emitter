package websocket

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryUpgradeNil(t *testing.T) {
	_, ok := TryUpgrade(nil, nil)
	assert.Equal(t, false, ok)
}

func TestTryUpgrade(t *testing.T) {
	//httptest.NewServer(handler)
	r := httptest.NewRequest("GET", "http://127.0.0.1/", bytes.NewBuffer([]byte{}))
	r.Header.Set("Connection", "upgrade")
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Sec-WebSocket-Extensions", "permessage-deflate; client_max_window_bits")
	r.Header.Set("Sec-WebSocket-Key", "D1icfJz+khA9kj5/14dRXQ==")
	r.Header.Set("Sec-WebSocket-Protocol", "mqttv3.1")
	r.Header.Set("Sec-WebSocket-Version", "13")

	w := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		TryUpgrade(w, r)
	})

	// TODO: need to have a hijackable response writer to test properly
	//ws, ok := TryUpgrade(w, r)
	//assert.NotNil(t, ws)
	//assert.True(t, ok)
}
