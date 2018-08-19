// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package dynamo

import (
	"context"
	"encoding/base64"

	"golang.org/x/crypto/acme/autocert"
)

// cache represents a certificate cache which uses AWS DynamoDB.
type cache struct {
	client *client
}

// newCache creates a new certificate cache.
func newCache(c *client) *cache {
	return &cache{
		client: c,
	}
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (p *cache) Get(ctx context.Context, key string) ([]byte, error) {
	if s, err := p.client.Get("certs/" + key); err == nil {
		return base64.StdEncoding.DecodeString(s)
	}
	return nil, autocert.ErrCacheMiss
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (p *cache) Put(ctx context.Context, key string, data []byte) error {
	return p.client.Put("certs/"+key, base64.StdEncoding.EncodeToString(data))
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (p *cache) Delete(ctx context.Context, key string) error {
	return p.client.Delete("certs/" + key)
}
