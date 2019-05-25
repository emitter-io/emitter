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

package usage

import (
	"errors"
	"testing"

	"github.com/emitter-io/emitter/internal/network/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoop_New(t *testing.T) {
	s := NewNoop()
	assert.Equal(t, &NoopStorage{}, s)
}

func TestNoop_Configure(t *testing.T) {
	s := new(NoopStorage)
	err := s.Configure(nil)
	assert.NoError(t, err)
}

func TestNoop_Name(t *testing.T) {
	s := new(NoopStorage)
	assert.Equal(t, "noop", s.Name())
}

func TestNoop_Get(t *testing.T) {
	s := new(NoopStorage)
	assert.Equal(t, uint32(123), s.Get(123).(Meter).GetContract())
}

func TestHTTP_New(t *testing.T) {
	s := NewHTTP()
	assert.NotNil(t, s.counters)

}

func TestHTTP_Name(t *testing.T) {
	s := new(HTTPStorage)
	assert.Equal(t, "http", s.Name())
}

func TestHTTP_ConfigureErr(t *testing.T) {
	s := NewHTTP()

	err := s.Configure(nil)
	assert.Error(t, errors.New("Configuration was not provided for HTTP metering provider"), err)

	err = s.Configure(map[string]interface{}{})
	assert.Error(t, errors.New("Configuration was not provided for HTTP metering provider"), err)
}

func TestHTTP_Configure(t *testing.T) {
	s := NewHTTP()

	err := s.Configure(map[string]interface{}{
		"interval":      1000.0,
		"url":           "http://localhost/test",
		"authorization": "test",
	})
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost/test", s.url)
	assert.NotNil(t, s.http)
	assert.NoError(t, s.Close())
}

func TestHTTP_Store(t *testing.T) {
	h := http.NewMockClient()
	h.On("Post", "http://127.0.0.1", mock.Anything, nil, mock.Anything).Return([]byte{}, nil)
	u1 := usage{MessageIn: 1, TrafficIn: 200, MessageEg: 1, TrafficEg: 100, Contract: 0x1}
	u2 := usage{MessageIn: 0, TrafficIn: 0, MessageEg: 0, TrafficEg: 0, Contract: 0x1}

	s := NewHTTP()
	s.url = "http://127.0.0.1"
	defer s.Close()
	s.http = h

	c := s.Get(1).(*usage)
	c.AddEgress(100)
	c.AddIngress(200)

	assert.Equal(t, u1.MessageIn, c.MessageIn)
	assert.Equal(t, u1.TrafficIn, c.TrafficIn)
	assert.Equal(t, u1.MessageEg, c.MessageEg)
	assert.Equal(t, u1.TrafficEg, c.TrafficEg)
	assert.Equal(t, u1.Contract, c.Contract)

	s.store()
	assert.Equal(t, u2.MessageIn, c.MessageIn)
	assert.Equal(t, u2.TrafficIn, c.TrafficIn)
	assert.Equal(t, u2.MessageEg, c.MessageEg)
	assert.Equal(t, u2.TrafficEg, c.TrafficEg)
	assert.Equal(t, u2.Contract, c.Contract)
}
