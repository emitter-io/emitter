package perf

import (
	"sync"
	"sync/atomic"
)

// Counter is an interface for integer increase only counter.
type Counter interface {
	Increment()
	IncrementBy(num int64)
	Decrement()
	DecrementBy(num int64)
	Reset()
	Name() string
	Value() int64
}

// Counters represents a container for various counters.
type Counters struct {
	sync.RWMutex
	counters map[string]Counter
}

// NewCounters creates a new object to keep all counters.
func NewCounters() *Counters {
	return &Counters{
		counters: make(map[string]Counter),
	}
}

// GetCounter returns a counter of given name, if doesn't exist than create.
func (c *Counters) GetCounter(name string) Counter {
	c.RLock()
	v, ok := c.counters[name]
	c.RUnlock()
	if !ok {
		c.Lock()
		if v, ok = c.counters[name]; !ok {
			v = &counterImpl{0, name}
			c.counters[name] = v
		}
		c.Unlock()
	}
	return v
}

// ------------------------------------------------------------------------------------

type counterImpl struct {
	value int64
	name  string
}

// Increment increases counter by one.
func (c *counterImpl) Increment() {
	atomic.AddInt64(&c.value, 1)
}

// IncrementBy increases counter by a given number.
func (c *counterImpl) IncrementBy(value int64) {
	atomic.AddInt64(&c.value, value)
}

// Decrement decreases counter by one.
func (c *counterImpl) Decrement() {
	atomic.AddInt64(&c.value, -1)
}

// DecrementBy decreases counter by a given number.
func (c *counterImpl) DecrementBy(value int64) {
	atomic.AddInt64(&c.value, -value)
}

// Name returns a name of counter.
func (c *counterImpl) Name() string {
	return c.name
}

// Reset sets the counter to zero.
func (c *counterImpl) Reset() {
	atomic.StoreInt64(&c.value, 0)
}

// Value returns a current value of counter.
func (c *counterImpl) Value() int64 {
	return atomic.LoadInt64(&c.value)
}
