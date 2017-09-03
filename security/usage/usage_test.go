package usage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUsageMeter(t *testing.T) {
	tracker := NewMeter(123)
	tracker.AddIngress(100)
	tracker.AddEgress(200)

	assert.Equal(t, uint32(123), tracker.GetContract())

	messageIn, trafficIn := tracker.GetIngress()
	messageEg, trafficEg := tracker.GetEgress()
	assert.Equal(t, int64(1), messageIn)
	assert.Equal(t, int64(100), trafficIn)

	assert.Equal(t, int64(1), messageEg)
	assert.Equal(t, int64(200), trafficEg)

	tracker.Reset()

	_, zero := tracker.GetIngress()
	assert.Equal(t, int64(0), zero)
}
