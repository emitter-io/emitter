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
	"sync/atomic"

	"github.com/axiomhq/hyperloglog"
)

// Meter represents a tracker for incoming and outgoing traffic.
type Meter interface {
	GetContract() uint32   // Returns the associated contract.
	AddIngress(size int64) // Records the ingress message size.
	AddEgress(size int64)  // Records the egress message size.
	AddDevice(addr string) // Records the device address.
}

// NewMeter constructs a new usage statistics instance.
func NewMeter(contract uint32) Meter {
	return &usage{
		Contract: contract,
		Devices:  hyperloglog.New(),
		Lock:     new(sync.Mutex),
	}
}

type usage struct {
	MessageIn int64 // Important to keep these here for alignment
	TrafficIn int64
	MessageEg int64
	TrafficEg int64
	Contract  uint32
	Lock      *sync.Mutex
	Devices   *hyperloglog.Sketch
}

// GetContract returns the associated contract.
func (t *usage) GetContract() uint32 {
	return t.Contract
}

// Records the ingress message size.
func (t *usage) AddIngress(size int64) {
	atomic.AddInt64(&t.MessageIn, 1)
	atomic.AddInt64(&t.TrafficIn, size)
}

// Records the egress message size.
func (t *usage) AddEgress(size int64) {
	atomic.AddInt64(&t.MessageEg, 1)
	atomic.AddInt64(&t.TrafficEg, size)
}

// Records the device address.
func (t *usage) AddDevice(addr string) {
	t.Lock.Lock()
	defer t.Lock.Unlock()
	t.Devices.Insert([]byte(addr))
}

// DeviceCount returns the estimated number of devices.
func (t *usage) DeviceCount() int {
	t.Lock.Lock()
	defer t.Lock.Unlock()
	return int(t.Devices.Estimate())
}

// reset resets the tracker and returns old usage.
func (t *usage) reset() encodedUsage {
	t.Lock.Lock()
	devices, _ := t.Devices.MarshalBinary()
	t.Devices = hyperloglog.New()
	t.Lock.Unlock()

	var old encodedUsage
	old.Contract = t.Contract
	old.MessageIn = atomic.SwapInt64(&t.MessageIn, 0)
	old.TrafficIn = atomic.SwapInt64(&t.TrafficIn, 0)
	old.MessageEg = atomic.SwapInt64(&t.MessageEg, 0)
	old.TrafficEg = atomic.SwapInt64(&t.TrafficEg, 0)
	old.Devices = devices
	return old
}

// merge merges in another usage.
func (t *usage) merge(other *usage) {
	t.Lock.Lock()
	t.Devices.Merge(other.Devices)
	t.Lock.Unlock()

	atomic.AddInt64(&t.MessageIn, other.MessageIn)
	atomic.AddInt64(&t.TrafficIn, other.TrafficIn)
	atomic.AddInt64(&t.MessageEg, other.MessageEg)
	atomic.AddInt64(&t.TrafficEg, other.TrafficEg)
}

// encodedUsage represents a single encoded usage which will be transfered
// within HTTP request and encoded into binary.
type encodedUsage struct {
	MessageIn int64
	TrafficIn int64
	MessageEg int64
	TrafficEg int64
	Contract  uint32
	Devices   []byte
}

// Converts the encoded usage to a normal usage.
func (t encodedUsage) toUsage() usage {
	d := hyperloglog.New()
	d.UnmarshalBinary(t.Devices)

	return usage{
		MessageIn: t.MessageIn,
		TrafficIn: t.TrafficIn,
		MessageEg: t.MessageEg,
		TrafficEg: t.TrafficEg,
		Contract:  t.Contract,
		Lock:      new(sync.Mutex),
		Devices:   d,
	}
}
