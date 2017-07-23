package perf

import (
	"sync"
	"sync/atomic"
)

// Counters represents a container for various counters.
type Counters struct {
	sync.RWMutex
	counters map[string]*Counter
}

// NewCounters creates a new object to keep all counters.
func NewCounters() *Counters {
	return &Counters{
		counters: make(map[string]*Counter),
	}
}

// GetCounter returns a counter of given name, if doesn't exist than create.
func (c *Counters) GetCounter(name string) *Counter {
	c.RLock()
	v, ok := c.counters[name]
	c.RUnlock()
	if !ok {
		c.Lock()
		if v, ok = c.counters[name]; !ok {
			v = &Counter{0, name}
			c.counters[name] = v
		}
		c.Unlock()
	}
	return v
}

// ------------------------------------------------------------------------------------

// Counter represents an atomic counter
type Counter struct {
	value int64
	name  string
}

// Increment increases counter by one.
func (c *Counter) Increment() {
	atomic.AddInt64(&c.value, 1)
}

// IncrementBy increases counter by a given number.
func (c *Counter) IncrementBy(value int64) {
	atomic.AddInt64(&c.value, value)
}

// Decrement decreases counter by one.
func (c *Counter) Decrement() {
	atomic.AddInt64(&c.value, -1)
}

// DecrementBy decreases counter by a given number.
func (c *Counter) DecrementBy(value int64) {
	atomic.AddInt64(&c.value, -value)
}

// Name returns a name of counter.
func (c *Counter) Name() string {
	return c.name
}

// Reset sets the counter to zero.
func (c *Counter) Reset() {
	atomic.StoreInt64(&c.value, 0)
}

// Value returns a current value of counter.
func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}
