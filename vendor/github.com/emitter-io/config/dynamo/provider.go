// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package dynamo

import (
	"errors"

	"golang.org/x/crypto/acme/autocert"
)

// Provider represents a security provider which uses hashicorp vault to store secrets.
type Provider struct {
	client *client
}

// NewProvider creates a new environment security provider.
func NewProvider() *Provider {
	return new(Provider)
}

// Name returns the name of the security provider.
func (p *Provider) Name() string {
	return "dynamodb"
}

// Configure configures the security provider.
func (p *Provider) Configure(config map[string]interface{}) (err error) {
	const prefix = "unable to configure DynamoDB provider"
	if config == nil {
		return errors.New(prefix + ", no configuration provided")
	}

	// Get AWS Region
	region := get(config, "region", "")
	if region == "" {
		return errors.New(prefix + ", no AWS Region was specified")
	}

	// Get AWS Table Name
	table := get(config, "table", "")
	if table == "" {
		return errors.New(prefix + ", no Table Name was specified")
	}

	// Get key and value columns, with defaults
	keyColumn := get(config, "keyColumn", "key")
	valColumn := get(config, "valueColumn", "value")

	// Create a new client
	p.client, err = newClient(region, table, keyColumn, valColumn)
	return err
}

// GetSecret retrieves a secret from the provider
func (p *Provider) GetSecret(secretName string) (string, bool) {
	if p.client != nil {
		if value, err := p.client.Get(secretName); err == nil {
			return value, true
		}
	}

	return "", false
}

// GetCache returns a certificate cache which can use the secrets store to read/write x509 certs.
func (p *Provider) GetCache() (autocert.Cache, bool) {
	if p.client != nil {
		return newCache(p.client), true
	}

	return nil, false
}

// Get retrieves a config value
func get(config map[string]interface{}, key, defaultValue string) string {
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}
