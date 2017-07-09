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

package collection

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const benchmarkItemCount = 1 << 10 // 1024

func setupConcurrentMap(b *testing.B) *ConcurrentMap {
	m := NewConcurrentMap()
	for i := uint32(0); i < benchmarkItemCount; i++ {
		m.Set(i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	return m
}

func setupGoMap(b *testing.B) map[uint32]interface{} {
	m := make(map[uint32]interface{})
	for i := uint32(0); i < benchmarkItemCount; i++ {
		m[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()
	return m
}

func BenchmarkRead_GoMap(b *testing.B) {
	m := setupGoMap(b)
	l := sync.Mutex{}

	for n := 0; n < b.N; n++ {
		i := n % benchmarkItemCount
		l.Lock()
		_, _ = m[uint32(i)]
		l.Unlock()
	}
}

func BenchmarkRead_ConcurrentMap(b *testing.B) {
	m := setupConcurrentMap(b)

	for n := 0; n < b.N; n++ {
		i := n % benchmarkItemCount
		m.Get(uint32(i))
	}
}

func BenchmarkParallelRead_GoMap(b *testing.B) {
	m := setupGoMap(b)
	l := sync.Mutex{}

	b.RunParallel(func(pb *testing.PB) {
		n := uint32(0)
		for pb.Next() {
			n++
			i := n % benchmarkItemCount
			l.Lock()
			_, _ = m[uint32(i)]
			l.Unlock()
		}
	})
}

func BenchmarkParallelRead_ConcurrentMap(b *testing.B) {
	m := setupConcurrentMap(b)

	b.RunParallel(func(pb *testing.PB) {
		n := uint32(0)
		for pb.Next() {
			n++
			i := n % benchmarkItemCount
			m.Get(uint32(i))
		}
	})
}

func BenchmarkWrite_ConcurrentMap(b *testing.B) {
	m := NewConcurrentMap()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		i := uint32(n) % benchmarkItemCount
		m.Set(i, i)
	}
}

func BenchmarkWrite_GoMap(b *testing.B) {
	m := make(map[uint32]interface{})
	l := &sync.Mutex{}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		i := uint32(n) % benchmarkItemCount

		l.Lock()
		m[i] = i
		l.Unlock()
	}
}

func TestConcurrentMap(t *testing.T) {
	value := "hi"

	m := NewConcurrentMap()
	m.Set(1, value)

	p, ok := m.Get(1)
	assert.True(t, ok)
	assert.Equal(t, value, p.(string))
}

func TestConcurrentMapRace(t *testing.T) {
	d := 2 * time.Second
	m := NewConcurrentMap()

	go testGoroutine(func() {
		m.Set(1, "hello")
	}, d)

	go testGoroutine(func() {
		m.Get(1)
	}, d)
}

func testGoroutine(f func(), d time.Duration) {
	for {
		select {
		case <-time.After(d):
			return
		default:
			f()
		}
	}
}
