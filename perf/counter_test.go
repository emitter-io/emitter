package perf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncrement(t *testing.T) {
	box := NewCounters()
	cnt := box.GetCounter("test")
	cnt.Increment()
	cnt.IncrementBy(7)
	assert.Equal(t, int64(8), cnt.Value())

	cnt.Decrement()
	cnt.DecrementBy(1)
	assert.Equal(t, int64(6), cnt.Value())
	assert.Equal(t, "test", cnt.Name())
	cnt.Reset()
	assert.Equal(t, int64(0), cnt.Value())
}

func TestIncrementParallel(t *testing.T) {
	box := NewCounters()
	end := make(chan bool, 10)
	for x := 0; x < 10; x++ {
		go func() {
			for y := 0; y < 100; y++ {
				cnt := box.GetCounter("test")
				cnt.Increment()
				cnt.IncrementBy(3)
			}
			end <- true
		}()
	}
	for i := 0; i < 10; {
		if _, ok := <-end; ok {
			i++
		}
	}

	if v := box.GetCounter("test"); v.Value() != 4000 {
		t.Errorf("got %d, expected 4000", v.Value())
	}
}

func BenchmarkCounters(b *testing.B) {
	b.StopTimer()
	e := make(chan bool)
	c := NewCounters()
	f := func(b *testing.B, c *Counters, e chan bool) {
		for i := 0; i < b.N; i++ {
			c.GetCounter("xxx").Increment()
		}
		e <- true
	}
	b.StartTimer()
	go f(b, c, e)
	go f(b, c, e)
	go f(b, c, e)
	go f(b, c, e)
	go f(b, c, e)
	go f(b, c, e)

	<-e
	<-e
	<-e
	<-e
	<-e
}

func BenchmarkCountersCached(b *testing.B) {
	b.StopTimer()
	e := make(chan bool)
	c := NewCounters()
	f := func(b *testing.B, c *Counters, e chan bool) {
		x := c.GetCounter("xxx")
		for i := 0; i < b.N; i++ {
			x.Increment()
		}
		e <- true
	}
	b.StartTimer()
	go f(b, c, e)
	go f(b, c, e)
	go f(b, c, e)
	go f(b, c, e)
	go f(b, c, e)
	go f(b, c, e)

	<-e
	<-e
	<-e
	<-e
	<-e
}
