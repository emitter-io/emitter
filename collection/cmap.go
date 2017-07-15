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
	"runtime"
	"sync/atomic"
)

// ConcurrentMap represents a thread-safe map.
type ConcurrentMap struct {
	lock int32
	read int32
	dict map[uint32]interface{}
}

// NewConcurrentMap creates a new thread-safe map.
func NewConcurrentMap() *ConcurrentMap {
	m := new(ConcurrentMap)
	m.dict = make(map[uint32]interface{})
	return m
}

// Get retrieves an element from the map.
func (m *ConcurrentMap) Get(key uint32) (v interface{}, ok bool) {
	atomic.AddInt32(&m.read, 1)
	for atomic.LoadInt32(&m.lock) > 0 {
		runtime.Gosched()
	}

	v, ok = m.dict[key]
	atomic.AddInt32(&m.read, -1)
	return
}

// GetOrCreate retrieves an element from the map or creates it using a factory function.
func (m *ConcurrentMap) GetOrCreate(key uint32, create func() interface{}) interface{} {
	atomic.AddInt32(&m.read, 1)
	for atomic.LoadInt32(&m.lock) > 0 {
		runtime.Gosched()
	}

	v, ok := m.dict[key]
	if !ok {
		if v = create(); v != nil {
			m.dict[key] = v
		}
	}

	atomic.AddInt32(&m.read, -1)
	return v
}

// Set sets an element to the map.
func (m *ConcurrentMap) Set(key uint32, value interface{}) {
	for !atomic.CompareAndSwapInt32(&m.lock, 0, 1) {
		runtime.Gosched()
	}

	for atomic.LoadInt32(&m.read) > 0 {
		runtime.Gosched()
	}

	m.dict[key] = value
	atomic.StoreInt32(&m.lock, 0)
}

// Delete deletes an element from the map.
func (m *ConcurrentMap) Delete(key uint32) {
	for !atomic.CompareAndSwapInt32(&m.lock, 0, 1) {
		runtime.Gosched()
	}

	for atomic.LoadInt32(&m.read) > 0 {
		runtime.Gosched()
	}

	delete(m.dict, key)
	atomic.StoreInt32(&m.lock, 0)
}
