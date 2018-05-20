// +build !js

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

package monitor

import (
	"runtime"
	"sync"
	"time"

	"github.com/golang/snappy"
	"github.com/kelindar/binary"
	"github.com/kelindar/process"
)

// Monitor represents a monitoring registry
type Monitor struct {
	registry sync.Map  // The registry used for keeping various metrics.
	created  time.Time // The start time for uptime calculation.
}

// Assert contract compliance
var _ Measurer = New()

// New creates a new monitor.
func New() *Monitor {
	return &Monitor{
		created: time.Now(),
	}
}

// Get retrieves a metric by its name. If the metric does not exist yet, it will
// create and register the metric.
func (m *Monitor) Get(name string) *Metric {
	if v, ok := m.registry.Load(name); ok {
		return v.(*Metric)
	}

	v, _ := m.registry.LoadOrStore(name, NewMetric(name))
	return v.(*Metric)
}

// Measure retrieves the metric and updates it.
func (m *Monitor) Measure(name string, value int32) {
	m.Get(name).Update(value)
}

// MeasureElapsed measures elapsed time since the start
func (m *Monitor) MeasureElapsed(name string, start time.Time) {
	m.Measure(name, int32(time.Since(start)/time.Microsecond))
}

// Tag updates a tag of a particular metric.
func (m *Monitor) Tag(name, tag string) {
	m.Get(name).UpdateTag(tag)
}

// Snapshot encodes the metrics into a binary representation
func (m *Monitor) Snapshot() (out []byte) {
	var snapshots []Snapshot
	m.registry.Range(func(k, v interface{}) bool {
		metric := v.(*Metric)
		snapshots = append(snapshots, *metric.Snapshot())
		metric.Reset()
		return true
	})

	// Marshal and compress with snappy
	if enc, err := binary.Marshal(snapshots); err == nil {
		out = snappy.Encode(out, enc)
	}
	return
}

// MeasureRuntime captures the runtime metrics, this is a relatively slow process
// and code is largely inspired by go-metrics.
func (m *Monitor) MeasureRuntime() {
	defer recover()

	// Collect stats
	var memory runtime.MemStats
	var memoryPriv, memoryVirtual int64
	var cpu float64
	runtime.ReadMemStats(&memory)
	process.ProcUsage(&cpu, &memoryPriv, &memoryVirtual)

	// Measure process information
	m.Measure("proc.cpu", int32(cpu*10000))
	m.Measure("proc.priv", int32(memoryPriv))
	m.Measure("proc.virt", int32(memoryVirtual))
	m.Measure("proc.uptime", int32(time.Now().Sub(m.created).Seconds()))

	// Measure heap information
	m.Measure("heap.alloc", int32(memory.HeapAlloc))
	m.Measure("heap.idle", int32(memory.HeapIdle))
	m.Measure("heap.inuse", int32(memory.HeapInuse))
	m.Measure("heap.objects", int32(memory.HeapObjects))
	m.Measure("heap.released", int32(memory.HeapReleased))
	m.Measure("heap.sys", int32(memory.HeapSys))

	// Measure off heap memory
	m.Measure("mcache.inuse", int32(memory.MCacheInuse))
	m.Measure("mcache.sys", int32(memory.MCacheSys))
	m.Measure("mspan.inuse", int32(memory.MSpanInuse))
	m.Measure("mspan.sys", int32(memory.MSpanSys))

	// Measure GC
	m.Measure("gc.cpu", int32(memory.GCCPUFraction*10000))
	m.Measure("gc.sys", int32(memory.GCSys))

	// Measure memory
	m.Measure("stack.inuse", int32(memory.StackInuse))
	m.Measure("stack.sys", int32(memory.StackSys))

	// Measure goroutines and threads and total memory
	m.Measure("go.count", int32(runtime.NumGoroutine()))
	m.Measure("go.procs", int32(runtime.NumCPU()))
	m.Measure("go.sys", int32(memory.Sys))
	m.Measure("go.alloc", int32(memory.TotalAlloc))
}
