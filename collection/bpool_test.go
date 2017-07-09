package collection

import (
	"bytes"
	"testing"
)

func TestBufferPool(t *testing.T) {
	size := 4
	capacity := 1024
	bufPool := NewBufferPool(size, capacity)
	b := bufPool.Get()

	// Check the cap before we use the buffer.
	if cap(b.Bytes()) != capacity {
		t.Fatalf("buffer capacity incorrect: got %v want %v", cap(b.Bytes()),
			capacity)
	}

	// Grow the buffer beyond our capacity and return it to the pool
	b.Grow(capacity * 3)
	bufPool.Put(b)

	// Add some additional buffers to fill up the pool.
	for i := 0; i < size; i++ {
		bufPool.Put(bytes.NewBuffer(make([]byte, 0, bufPool.size*2)))
	}

	// Check that oversized buffers are being replaced.
	if len(bufPool.pool) < size {
		t.Fatalf("buffer pool too small: got %v want %v", len(bufPool.pool), size)
	}

	// Close the channel so we can iterate over it.
	close(bufPool.pool)

	// Check that there are buffers of the correct capacity in the pool.
	for buffer := range bufPool.pool {
		if cap(buffer.Bytes()) != bufPool.size {
			t.Fatalf("returned buffers wrong capacity: got %v want %v",
				cap(buffer.Bytes()), capacity)
		}
	}

}
