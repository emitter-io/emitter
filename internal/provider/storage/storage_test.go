/**********************************************************************************
* Copyright (c) 2009-2018 Misakai Ltd.
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
	"fmt"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
)

type survey func(string, []byte) (message.Awaiter, error)

func (s survey) Survey(q string, b []byte) (message.Awaiter, error) {
	return s(q, b)
}

func testMessage(a, b, c uint32) *message.Message {
	return &message.Message{
		ID:      message.NewID(message.Ssid{0, a, b, c}),
		Channel: []byte("test/channel/"),
		Payload: []byte(fmt.Sprintf("%v,%v,%v", a, b, c)),
		TTL:     100,
	}
}

func TestNoop_Store(t *testing.T) {
	s := NewNoop()
	err := s.Store(testMessage(1, 2, 3))
	assert.NoError(t, err)
}

func TestNoop_Query(t *testing.T) {
	s := new(Noop)
	zero := time.Unix(0, 0)
	r, err := s.Query(testMessage(1, 2, 3).Ssid(), zero, zero, 10)
	assert.NoError(t, err)
	for range r {
		t.Errorf("Should be empty")
	}
}

func TestNoop_Configure(t *testing.T) {
	s := new(Noop)
	err := s.Configure(nil)
	assert.NoError(t, err)
}

func TestNoop_Name(t *testing.T) {
	s := new(Noop)
	assert.Equal(t, "noop", s.Name())
}

func TestNoop_Close(t *testing.T) {
	s := new(Noop)
	err := s.Close()
	assert.NoError(t, err)
}
