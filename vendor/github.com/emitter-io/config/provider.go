package config

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// ------------------------------------------------------------------------------------

// EnvironmentProvider represents a security provider which uses environment variables to store secrets.
type EnvironmentProvider struct {
	lookup func(string) (string, bool)
}

// NewEnvironmentProvider creates a new environment security provider.
func NewEnvironmentProvider() *EnvironmentProvider {
	return &EnvironmentProvider{
		lookup: os.LookupEnv,
	}
}

// Configure configures the security provider.
func (p *EnvironmentProvider) Configure(c Config) error {
	return nil
}

// GetSecret retrieves a secret from the provider
func (p *EnvironmentProvider) GetSecret(secretName string) (string, bool) {
	name := strings.ToUpper(strings.Replace(secretName, "/", "_", -1))
	return p.lookup(name)
}

// ------------------------------------------------------------------------------------

// VaultProvider represents a security provider which uses hashicorp vault to store secrets.
type VaultProvider struct {
	client *VaultClient // The vault client.
	app    string       // The application ID to use for authentication.
	user   string       // The user ID to use for authentication.
	auth   bool         // Whether the provider is authenticated or not.
}

// NewVaultProvider creates a new environment security provider.
func NewVaultProvider(user string) *VaultProvider {
	return &VaultProvider{user: user}
}

// Configure configures the security provider.
func (p *VaultProvider) Configure(c Config) error {
	if c.Vault() == nil || c.Vault().Address == "" || c.Vault().Application == "" {
		return errors.New("unable to configure Vault provider")
	}

	// Create a new client
	cli, err := c.Vault().NewClient(p.user)
	p.client = cli
	return err
}

// GetSecret retrieves a secret from the provider
func (p *VaultProvider) GetSecret(secretName string) (string, bool) {
	if p.client != nil && p.client.IsAuthenticated() {
		if value, err := p.client.ReadSecret(secretName); err == nil {
			return value, true
		}
	}

	return "", false
}

// ------------------------------------------------------------------------------------

// VaultCache represents a certificate cache which uses hashicorp vault.
type VaultCache struct {
	client *VaultClient // The vault client.
}

// NewVaultCache creates a new certificate cache.
func NewVaultCache(user string, c Config) (*VaultCache, error) {
	if c.Vault() == nil || c.Vault().Address == "" || c.Vault().Application == "" {
		return nil, errors.New("unable to configure Vault certificate cache")
	}

	cli, err := c.Vault().NewClient(user)
	if err != nil {
		return nil, err
	}

	return &VaultCache{
		client: cli,
	}, nil
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (p *VaultCache) Get(ctx context.Context, key string) ([]byte, error) {
	s, err := p.client.ReadSecret("certs/" + key)
	if err != nil {
		return nil, autocert.ErrCacheMiss
	}

	return base64.StdEncoding.DecodeString(s)
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (p *VaultCache) Put(ctx context.Context, key string, data []byte) error {
	return p.client.WriteSecret("certs/"+key, base64.StdEncoding.EncodeToString(data))
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (p *VaultCache) Delete(ctx context.Context, key string) error {
	return p.client.WriteSecret("certs/"+key, "")
}
