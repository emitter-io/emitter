package usage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUsageMeter(t *testing.T) {
	tracker := NewMeter(123)
	assert.Equal(t, uint32(123), tracker.GetContract())
}

func TestMeterAdd(t *testing.T) {
	tracker := &usage{Contract: 123}
	tracker.AddIngress(100)
	tracker.AddEgress(200)

	assert.Equal(t, uint32(123), tracker.GetContract())

	assert.Equal(t, int64(1), tracker.MessageIn)
	assert.Equal(t, int64(100), tracker.TrafficIn)

	assert.Equal(t, int64(1), tracker.MessageEg)
	assert.Equal(t, int64(200), tracker.TrafficEg)
}

func TestMeterReset(t *testing.T) {
	meter := &usage{TrafficIn: 1000}
	old := meter.reset()

	assert.Equal(t, int64(1000), old.TrafficIn)
	assert.Equal(t, int64(0), meter.TrafficIn)
}
