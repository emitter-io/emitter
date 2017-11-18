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
	"testing"

	"github.com/emitter-io/emitter/broker/message"
	"github.com/stretchr/testify/assert"
)

func TestHTTP_Name(t *testing.T) {
	s := NewHTTP()
	assert.Equal(t, "http", s.Name())
}

func TestHTTP_Configure(t *testing.T) {
	s := NewHTTP()
	cfg := map[string]interface{}{
		"interval": float64(100),
		"url":      "http://127.0.0.1/",
	}

	err := s.Configure(cfg)
	assert.NoError(t, err)

	errClose := s.Close()
	assert.NoError(t, errClose)
}

func TestHTTP_format(t *testing.T) {
	s := NewHTTP()

	assert.Equal(t, "msg/append", s.buildAppendURL())
	assert.Equal(t, "msg/last?ssid=[1,2,3]&n=100", s.buildLastURL([]uint32{1, 2, 3}, 100))
}

func TestHTTP_Store(t *testing.T) {
	s := NewHTTP()

	s.Store(&message.Message{
		Time: 0,
	})
	assert.Equal(t, 1, len(s.frame))
}
