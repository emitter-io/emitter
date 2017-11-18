package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/emitter-io/emitter/utils"
	"github.com/stretchr/testify/assert"
)

type testHandler struct{}

type testObject struct {
	Field string `json:"field"`
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		url string
		ok  bool
	}{
		{url: "http://google.com/123", ok: true},
		{url: "google.com/123", ok: false},
		{url: "235235", ok: false},
		{url: "::", ok: false},
	}

	for _, tc := range tests {
		c, err := NewClient(tc.url, time.Second)
		if tc.ok {
			assert.NotNil(t, c)
			assert.NoError(t, err)
		} else {
			assert.Nil(t, c)
		}
	}
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response []byte
	if r.Header.Get("Content-Type") == "application/binary" {
		w.Header().Set("Content-Type", "application/binary")
		response, _ = utils.Encode(&testObject{
			Field: "response",
		})
	} else {
		w.Header().Set("Content-Type", "application/json")
		response, _ = json.Marshal(&testObject{
			Field: "response",
		})
	}

	w.Write(response)
}

func TestPostGet(t *testing.T) {
	s := httptest.NewServer(new(testHandler))
	defer s.Close()
	body := testObject{Field: "hello"}
	expect := &testObject{Field: "response"}

	jsonBody, _ := json.Marshal(body)

	// Reuse the client
	c, err := NewClient(s.URL, time.Second, NewHeader("Authorization", "123"))
	assert.NoError(t, err)

	{
		output := new(testObject)
		err := c.Get(s.URL, output)
		assert.NoError(t, err)
		assert.EqualValues(t, expect, output)
	}

	{
		output := new(testObject)
		err := c.Post(s.URL, jsonBody, output)
		assert.NoError(t, err)
		assert.EqualValues(t, expect, output)
	}

	{
		output := new(testObject)
		err := c.PostJSON(s.URL, body, output)
		assert.NoError(t, err)
		assert.EqualValues(t, expect, output)
	}

	{
		output := new(testObject)
		err := c.PostBinary(s.URL, body, output)
		assert.NoError(t, err)
		assert.EqualValues(t, expect, output)
	}
}
