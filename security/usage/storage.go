package usage

import (
	"github.com/emitter-io/config"
)

// Metering represents a contract for a usage metering
type Metering interface {
	config.Provider

	// Store stores the meters in some underlying usage storage.
	Store() error

	// Get retrieves a meter for a contract.
	Get(id uint32) Meter
}

// ------------------------------------------------------------------------------------

// Noop implements Storage contract.
var _ Metering = new(NoopStorage)

// NoopStorage represents a usage storage which does nothing.
type NoopStorage struct{}

// Name returns the name of the provider.
func (s *NoopStorage) Name() string {
	return "noop"
}

// Configure configures the provider
func (s *NoopStorage) Configure(config map[string]interface{}) error {
	return nil
}

// Store stores the meters in some underlying usage storage.
func (s *NoopStorage) Store() error {
	return nil
}

// Get retrieves a meter for a contract..
func (s *NoopStorage) Get(id uint32) Meter {
	return NewMeter(id)
}
