// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package vault

import (
	"errors"

	"golang.org/x/crypto/acme/autocert"
)

// Provider represents a security provider which uses hashicorp vault to store secrets.
type Provider struct {
	client *client // The vault client.
	app    string  // The application ID to use for authentication.
	user   string  // The user ID to use for authentication.
	auth   bool    // Whether the provider is authenticated or not.
}

// NewProvider creates a new environment security provider.
func NewProvider(user string) *Provider {
	return &Provider{user: user}
}

// Name returns the name of the security provider.
func (p *Provider) Name() string {
	return "vault"
}

// Configure configures the security provider.
func (p *Provider) Configure(config map[string]interface{}) (err error) {
	if config == nil {
		return errors.New("unable to configure Vault provider, no configuration provided")
	}

	// Get configuration
	address, _ := config["address"]
	app, _ := config["app"]
	if address == "" || app == "" {
		return errors.New("unable to configure Vault provider, no address or app provided")
	}

	// Create a new client
	cli := newClient(address.(string))
	err = cli.Authenticate(app.(string), p.user)
	p.client = cli
	return err
}

// GetSecret retrieves a secret from the provider
func (p *Provider) GetSecret(secretName string) (string, bool) {
	if p.client != nil && p.client.IsAuthenticated() {
		if value, err := p.client.ReadSecret(secretName); err == nil {
			return value, true
		}
	}

	return "", false
}

// GetCache returns a certificate cache which can use the secrets store to read/write x509 certs.
func (p *Provider) GetCache() (autocert.Cache, bool) {
	if p.client != nil && p.client.IsAuthenticated() {
		return newCache(p.client), true
	}

	return nil, false
}
