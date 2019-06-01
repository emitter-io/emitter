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

package mqtt

import (
	"sync"
)

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64
const maxInt = int(^uint(0) >> 1)

// buffers are reusable fixed-side buffers for faster encoding.
var buffers = newBufferPool(maxMessageSize)

// bufferPool represents a thread safe buffer pool
type bufferPool struct {
	sync.Pool
}

// newBufferPool creates a new BufferPool bounded to the given size.
func newBufferPool(bufferSize int) (bp *bufferPool) {
	return &bufferPool{
		sync.Pool{
			New: func() interface{} {
				return &byteBuffer{buf: make([]byte, bufferSize)}
			},
		},
	}
}

// Get gets a Buffer from the SizedBufferPool, or creates a new one if none are
// available in the pool. Buffers have a pre-allocated capacity.
func (bp *bufferPool) Get() *byteBuffer {
	return bp.Pool.Get().(*byteBuffer)
}

// Put returns the given Buffer to the SizedBufferPool.
func (bp *bufferPool) Put(b *byteBuffer) {
	bp.Pool.Put(b)
}

type byteBuffer struct {
	buf []byte
}

// Bytes gets a byte slice of a specified size.
func (b *byteBuffer) Bytes(n int) []byte {
	if n == 0 { // Return max size
		return b.buf
	}

	return b.buf[:n]
}

// Slice returns a slice at an offset.
func (b *byteBuffer) Slice(from, until int) []byte {
	return b.buf[from:until]
}

// Split splits the bufer in two.
func (b *byteBuffer) Split(n int) ([]byte, []byte) {
	buffer := b.Bytes(0)
	return buffer[:n], buffer[n:]
}
