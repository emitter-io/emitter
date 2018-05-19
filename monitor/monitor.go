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
	"context"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/golang/snappy"
	"github.com/kelindar/binary"
)

// Measurer represents a monitoring contract.
type Measurer interface {
	MeasureValue(name string, value int64)
	MeasureElapsed(name string, start time.Time)
}

// Snapshotter represents a snapshotting contract.
type Snapshotter interface {
	SnapshotSink(ctx context.Context, interval time.Duration, sink io.Writer)
	Snapshot() []byte
}

// Monitor represents a monitoring registry
type Monitor struct {
	registry sync.Map
}

// Assert contract compliance
var _ Measurer = New()
var _ Snapshotter = New()

// New creates a new monitor.
func New() *Monitor {
	return new(Monitor)
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

// MeasureValue retrieves the metric and updates it.
func (m *Monitor) MeasureValue(name string, value int64) {
	m.Get(name).Update(value)
}

// MeasureElapsed measures elapsed time since the start
func (m *Monitor) MeasureElapsed(name string, start time.Time) {
	m.MeasureValue(name, int64(time.Since(start)/time.Millisecond))
}

// SnapshotSink performs snapshots asynchronously and keeps pushing it into the writer.
func (m *Monitor) SnapshotSink(ctx context.Context, interval time.Duration, sink io.Writer) {
	timer := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				m.MeasureRuntime()
				snap := m.Snapshot()
				println("snapshot size", len(snap))
				sink.Write(snap)
			}
		}
	}()
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
	memStats := new(runtime.MemStats)
	runtime.ReadMemStats(memStats)

	// Measure heap information
	m.MeasureValue("heap.alloc", int64(memStats.HeapAlloc))
	m.MeasureValue("heap.idle", int64(memStats.HeapIdle))
	m.MeasureValue("heap.inuse", int64(memStats.HeapInuse))
	m.MeasureValue("heap.objects", int64(memStats.HeapObjects))
	m.MeasureValue("heap.released", int64(memStats.HeapReleased))
	m.MeasureValue("heap.sys", int64(memStats.HeapSys))

	// Measure off heap memory
	m.MeasureValue("mcache.inuse", int64(memStats.MCacheInuse))
	m.MeasureValue("mcache.sys", int64(memStats.MCacheSys))
	m.MeasureValue("mspan.inuse", int64(memStats.MSpanInuse))
	m.MeasureValue("mspan.sys", int64(memStats.MSpanSys))

	// Measure GC
	m.MeasureValue("gc.cpu", int64(memStats.GCCPUFraction*10000))
	m.MeasureValue("gc.sys", int64(memStats.GCSys))

	// Measure memory
	m.MeasureValue("stack.inuse", int64(memStats.StackInuse))
	m.MeasureValue("stack.sys", int64(memStats.StackSys))
	m.MeasureValue("mem.sys", int64(memStats.Sys))
	m.MeasureValue("mem.alloc", int64(memStats.TotalAlloc))

	// Measure goroutines and threads
	m.MeasureValue("go.count", int64(runtime.NumGoroutine()))
	m.MeasureValue("go.procs", int64(runtime.NumCPU()))
}

// Restore restores a snapshot into a read-only histogram format.
func Restore(encoded []byte) (snapshots []Snapshot, err error) {
	var decoded []byte
	if decoded, err = snappy.Decode(decoded, encoded); err == nil {
		err = binary.Unmarshal(decoded, &snapshots)
	}
	return
}
