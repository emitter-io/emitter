/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
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

package storage

import (
	"errors"
	netHttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/emitter-io/emitter/network/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHTTP_Name(t *testing.T) {
	s := NewHTTP()
	assert.Equal(t, "http", s.Name())
}

func TestHTTP_ConfigureNil(t *testing.T) {
	s := NewHTTP()

	err1 := s.Configure(nil)
	assert.Error(t, err1)

	err2 := s.Configure(map[string]interface{}{})
	assert.Error(t, err2)

	errClose := s.Close()
	assert.NoError(t, errClose)
}

func TestHTTP_Configure(t *testing.T) {
	s := NewHTTP()
	cfg := map[string]interface{}{
		"interval":      float64(100),
		"url":           "http://127.0.0.1/",
		"authorization": "Digest 1234",
	}

	err := s.Configure(cfg)
	assert.NoError(t, err)

	errClose := s.Close()
	assert.NoError(t, errClose)
}
func TestHTTP_format(t *testing.T) {
	s := NewHTTP()

	assert.Equal(t, "v1/add/", s.buildAppendURL())
	assert.Equal(t, "v1/get/?ssid=[1,2,3]&limit=100", s.buildLastURL([]uint32{1, 2, 3}, 100))
}

func TestHTTP_Store(t *testing.T) {
	h := http.NewMockClient()
	h.On("Post", "v1/add/", mock.Anything, nil, mock.Anything).Return([]byte{}, nil).Once()

	s := NewHTTP()
	s.http = h

	s.Store(&message.Message{})
	assert.Equal(t, 1, len(s.frame))

	s.store()
	assert.Equal(t, 0, len(s.frame))
}

func TestHTTP_StoreError(t *testing.T) {
	h := http.NewMockClient()
	h.On("Post", "v1/add/", mock.Anything, nil, mock.Anything).Return([]byte{}, errors.New("boom")).Once()

	s := NewHTTP()
	s.http = h

	s.Store(&message.Message{})
	assert.Equal(t, 1, len(s.frame))

	// We should have pushed the frame back
	s.store()
	assert.Equal(t, 1, len(s.frame))
}

func TestHTTP_QueryLast(t *testing.T) {
	frame := message.Frame{
		*testMessage(1, 2, 3),
		*testMessage(1, 2, 3),
	}

	encoded := frame.Encode()

	h := http.NewMockClient()
	h.On("Get", "v1/get/?ssid=[1,2,3]&limit=10", nil, mock.Anything).Return(encoded, nil).Once()

	s := NewHTTP()
	s.http = h

	out, err := s.QueryLast([]uint32{1, 2, 3}, 10)
	assert.NoError(t, err)

	count := 0
	for range out {
		count++
	}

	assert.Equal(t, 2, count)

	h.On("Get", "v1/get/?ssid=[1,2,3]&limit=10", nil, mock.Anything).Return(encoded, errors.New("boom")).Once()
	_, err = s.QueryLast([]uint32{1, 2, 3}, 10)
	assert.Error(t, err)

}

type handler1 struct {
	url string
}

func (h *handler1) ServeHTTP(w netHttp.ResponseWriter, r *netHttp.Request) {
	w.Header().Set("Location", h.url+r.URL.String())
	w.WriteHeader(308)
}

type handler2 struct{}

func (h *handler2) ServeHTTP(w netHttp.ResponseWriter, r *netHttp.Request) {
	frame := message.Frame{
		*testMessage(1, 2, 3),
		*testMessage(1, 2, 3),
		*testMessage(1, 2, 3),
	}

	encoded := frame.Encode()
	w.Write(encoded)
	w.WriteHeader(200)
}

func TestHTTP_Redirect(t *testing.T) {
	handler1 := new(handler1)
	server1 := httptest.NewServer(handler1)
	server2 := httptest.NewServer(new(handler2))
	handler1.url = server2.URL
	defer server1.Close()
	defer server2.Close()

	s := NewHTTP()
	s.Configure(map[string]interface{}{
		"interval": float64(1),
		"url":      server1.URL,
	})

	out, err := s.QueryLast([]uint32{1, 2, 3}, 10)
	assert.NoError(t, err)
	count := 0

	for range out {
		count++
	}

	assert.Equal(t, 3, count)
}
