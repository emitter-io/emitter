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
	"bytes"
)

// BufferPool represents a thread safe buffer pool
type BufferPool struct {
	pool chan *bytes.Buffer
	size int
}

// NewBufferPool creates a new BufferPool bounded to the given size.
func NewBufferPool(initialCapacity int, bufferSize int) (bp *BufferPool) {
	return &BufferPool{
		pool: make(chan *bytes.Buffer, initialCapacity),
		size: bufferSize,
	}
}

// Get gets a Buffer from the SizedBufferPool, or creates a new one if none are
// available in the pool. Buffers have a pre-allocated capacity.
func (bp *BufferPool) Get() (b *bytes.Buffer) {
	select {
	case b = <-bp.pool: // reuse existing buffer
	default:
		// create new buffer
		b = bytes.NewBuffer(make([]byte, 0, bp.size))
	}
	return
}

// Put returns the given Buffer to the SizedBufferPool.
func (bp *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()

	// Release buffers over our maximum capacity and re-create a pre-sized
	// buffer to replace it.
	if cap(b.Bytes()) > bp.size {
		b = bytes.NewBuffer(make([]byte, 0, bp.size))
	}

	select {
	case bp.pool <- b:
	default: // Discard the buffer if the pool is full.
	}
}
