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
	"sync"
	"testing"

	"github.com/axiomhq/hyperloglog"
	"github.com/stretchr/testify/assert"
)

func TestNewUsageMeter(t *testing.T) {
	meter := NewMeter(123)
	assert.Equal(t, uint32(123), meter.GetContract())
}

func TestMeterAdd(t *testing.T) {
	meter := &usage{Contract: 123, Lock: new(sync.Mutex), Devices: hyperloglog.New()}
	meter.AddIngress(100)
	meter.AddEgress(200)
	meter.AddDevice("123")
	meter.AddDevice("456")

	assert.Equal(t, uint32(123), meter.GetContract())

	assert.Equal(t, int64(1), meter.MessageIn)
	assert.Equal(t, int64(100), meter.TrafficIn)

	assert.Equal(t, int64(1), meter.MessageEg)
	assert.Equal(t, int64(200), meter.TrafficEg)

	assert.Equal(t, uint64(2), meter.Devices.Estimate())
}

func TestMeterReset(t *testing.T) {
	meter := &usage{Lock: new(sync.Mutex), Devices: hyperloglog.New()}

	// Add a device and some traffic
	meter.AddIngress(1000)
	meter.AddDevice("123")
	meter.AddDevice("153")
	old1 := meter.reset().toUsage()

	// Assert
	assert.Equal(t, int64(1000), old1.TrafficIn)
	assert.Equal(t, int64(0), meter.TrafficIn)
	assert.Equal(t, 0, meter.DeviceCount())
	assert.Equal(t, 2, old1.DeviceCount())

	// Add a device and some traffic and reset again
	meter.AddIngress(1000)
	meter.AddDevice("123")
	meter.AddDevice("345")
	old2 := meter.reset().toUsage()

	// Assert
	assert.Equal(t, int64(1000), old2.TrafficIn)
	assert.Equal(t, int64(0), meter.TrafficIn)
	assert.Equal(t, 0, meter.DeviceCount())
	assert.Equal(t, 2, old2.DeviceCount())

	// Merge in
	old1.merge(&old2)
	assert.Equal(t, int64(2000), old1.TrafficIn)
	assert.Equal(t, 3, old1.DeviceCount())
}
