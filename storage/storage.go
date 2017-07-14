package storage

import (
	"os"

	"github.com/dgraph-io/badger"
)

// Storage represents a durable key/value store used internally by the broker.
type Storage *badger.KV

// New creates a new storage instance.
func New() Storage {
	os.Mkdir("./data", os.ModePerm)

	opts := badger.DefaultOptions
	opts.Dir = "./data"
	opts.ValueDir = "./data"

	kv, err := badger.NewKV(&opts)
	if err != nil {
		return nil
	}

	return Storage(kv)
}
