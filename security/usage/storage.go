package usage

import (
	"errors"
	"github.com/emitter-io/emitter/logging"
	"sync"
	"time"

	"github.com/emitter-io/config"
	"github.com/emitter-io/emitter/network/http"
	"github.com/emitter-io/emitter/utils"
)

// Metering represents a contract for a usage metering
type Metering interface {
	config.Provider

	// Store stores the meters in some underlying usage storage.
	Store() error

	// Get retrieves a meter for a contract.
	Get(id uint32) interface{}
}

// ------------------------------------------------------------------------------------

// Noop implements Storage contract.
var _ Metering = new(NoopStorage)

// NoopStorage represents a usage storage which does nothing.
type NoopStorage struct{}

// NewNoop creates a new no-op storage.
func NewNoop() *NoopStorage {
	return new(NoopStorage)
}

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

// Get retrieves a meter for a contract.
func (s *NoopStorage) Get(id uint32) interface{} {
	return NewMeter(id)
}

// ------------------------------------------------------------------------------------

// HTTPStorage represents a usage storage which posts meters over HTTP.
type HTTPStorage struct {
	counters *sync.Map
	url      string
}

// NewHTTP creates a new HTTP storage
func NewHTTP() *HTTPStorage {
	return &HTTPStorage{
		counters: new(sync.Map),
	}
}

// Name returns the name of the provider.
func (s *HTTPStorage) Name() string {
	return "http"
}

// Configure configures the provider.
func (s *HTTPStorage) Configure(config map[string]interface{}) error {
	if config == nil {
		return errors.New("Configuration was not provided for HTTP metering provider")
	}

	// Get the interval from the provider configuration
	interval := time.Second
	if v, ok := config["interval"]; ok {
		if i, ok := v.(float64); ok {
			interval = time.Duration(i) * time.Millisecond
		}
	}

	// Get the url from the provider configuration
	if url, ok := config["url"]; ok {
		s.url = url.(string)
		utils.Repeat(s.store, interval, make(chan bool)) // TODO: closing chan
		return nil
	}

	return errors.New("The 'url' parameter was not provider in the configuration for HTTP contract provider")
}

// Get retrieves a meter for a contract.
func (s *HTTPStorage) Get(id uint32) interface{} {
	meter, _ := s.counters.LoadOrStore(id, NewMeter(id))
	return meter
}

// Store periodically stores the counters by sending them through HTTP.
func (s *HTTPStorage) store() {
	counters := make([]*usage, 0)
	s.counters.Range(func(k, v interface{}) bool {
		counters = append(counters, v.(*usage).reset())
		return true
	})

	var out interface{}
	if err := http.Post(s.url, counters, &out); err != nil {
		logging.LogError("http metering", "reporting counters", err)
	}
}
