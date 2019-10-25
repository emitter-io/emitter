/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/

package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
)

type handler func(http.ResponseWriter, *http.Request)

func (f handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

type testHandler struct{}

type testObject struct {
	Field string `json:"field"`
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		url string
	}{
		{url: "http://google.com/123"},
		{url: "google.com/123"},
	}

	for _, tc := range tests {
		c, err := NewHostClient(tc.url, time.Second)
		assert.NotNil(t, c)
		assert.NoError(t, err)
	}
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response []byte
	if r.Header.Get("Content-Type") == "application/binary" {
		w.Header().Set("Content-Type", "application/binary")
		response, _ = binary.Marshal(&testObject{
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
	c, err := NewHostClient(s.URL, time.Second, NewHeader("Authorization", "123"))
	assert.NoError(t, err)

	{
		output := new(testObject)
		_, err := c.Get(s.URL, output)
		assert.NoError(t, err)
		assert.EqualValues(t, expect, output)
	}

	{
		body, err := c.Get(s.URL, nil)
		assert.NoError(t, err)
		assert.NotNil(t, body)
	}

	{
		output := new(testObject)
		_, err := c.Post(s.URL, jsonBody, output)
		assert.NoError(t, err)
		assert.EqualValues(t, expect, output)
	}

}

type handler1 struct {
	url string
}

func (h *handler1) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", h.url)
	w.WriteHeader(308)
}

type handler2 struct{}

func (h *handler2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response []byte
	w.Header().Set("Content-Type", "application/binary")
	response, _ = binary.Marshal(&testObject{
		Field: "response",
	})
	w.Write(response)
	w.WriteHeader(200)
}

func TestHTTP_Redirect(t *testing.T) {
	handler1 := new(handler1)
	server1 := httptest.NewServer(handler1)
	server2 := httptest.NewServer(new(handler2))
	handler1.url = server2.URL
	defer server1.Close()
	defer server2.Close()

	// New client
	c, err := NewHostClient(server1.URL, time.Second)
	assert.NoError(t, err)

	// Get something from server1
	output := new(testObject)
	_, err = c.Get(server1.URL, output, NewHeader("X-Test-Header", "123"))
	assert.NoError(t, err)
}

func TestHTTP_204(t *testing.T) {
	server := httptest.NewServer(handler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	// New client
	c, err := NewClient(time.Second)
	b, err := c.Get(server.URL, nil)
	assert.NoError(t, err)
	assert.Nil(t, b)
}

func TestHTTP_500(t *testing.T) {
	server := httptest.NewServer(handler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	// New client
	c, err := NewClient(time.Second)
	b, err := c.Get(server.URL, nil)
	assert.Error(t, err)
	assert.Nil(t, b)
}
