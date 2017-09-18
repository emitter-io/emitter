package collection

import (
	"testing"
)

func TestBufferPool(t *testing.T) {
	capacity := 1024
	bufPool := NewBufferPool(capacity)
	b := bufPool.Get()

	// Check the cap before we use the buffer.
	if cap(b.Bytes()) != capacity {
		t.Fatalf("buffer capacity incorrect: got %v want %v", cap(b.Bytes()),
			capacity)
	}

	// Grow the buffer beyond our capacity and return it to the pool
	b.Grow(capacity * 3)
	bufPool.Put(b)

}

func BenchmarkBufferPool(b *testing.B) {
	pool := NewBufferPool(100)

	for n := 0; n < b.N; n++ {
		b := pool.Get()
		pool.Put(b)
	}
}
