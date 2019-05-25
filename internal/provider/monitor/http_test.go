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

package monitor

import (
	"io/ioutil"
	netHttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type handler func(netHttp.ResponseWriter, *netHttp.Request)

func (f handler) ServeHTTP(w netHttp.ResponseWriter, r *netHttp.Request) {
	f(w, r)
}

func TestHTTP_HappyPath(t *testing.T) {
	r := snapshot("test")
	server := httptest.NewServer(handler(func(w netHttp.ResponseWriter, r *netHttp.Request) {
		b, err := ioutil.ReadAll(r.Body)
		assert.Equal(t, "test", string(b))
		assert.NoError(t, err)
		w.WriteHeader(204)
	}))
	defer server.Close()

	s := NewHTTP(r)
	err := s.Configure(map[string]interface{}{
		"interval":      float64(100),
		"url":           server.URL,
		"authorization": "123",
	})

	assert.NoError(t, err)
	assert.Equal(t, "http", s.Name())

	errClose := s.Close()
	assert.NoError(t, errClose)
}

func TestHTTP_ErrorPost(t *testing.T) {
	r := snapshot("test")
	server := httptest.NewServer(handler(func(w netHttp.ResponseWriter, r *netHttp.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	s := NewHTTP(r)
	err := s.Configure(map[string]interface{}{
		"interval": float64(100),
		"url":      server.URL,
	})

	assert.NoError(t, err)
	assert.Equal(t, "http", s.Name())

	errClose := s.Close()
	assert.NoError(t, errClose)
}

func TestHTTP_ErrorConfig(t *testing.T) {
	{
		s := NewHTTP(nil)
		err := s.Configure(nil)
		assert.Error(t, err)
	}
	{
		s := NewHTTP(nil)
		err := s.Configure(map[string]interface{}{})
		assert.Error(t, err)
	}
}
