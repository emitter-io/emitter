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
	"os"
	"testing"

	"github.com/emitter-io/stats"
	"github.com/stretchr/testify/assert"
)

func TestStatsd_HappyPath(t *testing.T) {
	m := stats.New()
	for i := int32(0); i < 100; i++ {
		m.Measure("proc.test", i)
		m.Measure("node.test", i)
		m.Measure("rcv.test", i)
	}

	s := NewStatsd(m, "")
	defer s.Close()

	s.Configure(map[string]interface{}{
		"interval": 1000000.00,
		"url":      ":8125",
	})
	assert.NotPanics(t, func() {
		s.write()
	})
}

func TestStatsd_BadSnapshot(t *testing.T) {
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		t.Skip("Skipping the test in CI environment")
		return
	}

	r := snapshot("test")
	s := NewStatsd(r, "")
	defer s.Close()

	err := s.Configure(map[string]interface{}{
		"interval": 1000000.00,
		"url":      ":8125",
	})
	assert.NoError(t, err)
	assert.NotPanics(t, func() {
		s.write()
	})
}

func TestStatsd_Configure(t *testing.T) {
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		t.Skip("Skipping the test in CI environment")
		return
	}

	{
		s := NewStatsd(nil, "")
		defer s.Close()
		assert.Equal(t, "statsd", s.Name())

		err := s.Configure(nil)
		assert.NoError(t, err)
	}

	{
		s := NewStatsd(nil, "")
		defer s.Close()
		assert.Equal(t, "statsd", s.Name())

		err := s.Configure(map[string]interface{}{})
		assert.NoError(t, err)
	}

	{
		s := NewStatsd(nil, "")
		defer s.Close()
		assert.Equal(t, "statsd", s.Name())

		err := s.Configure(map[string]interface{}{
			"interval": 100.00,
			"url":      ":8125",
		})
		assert.NoError(t, err)
	}
}
