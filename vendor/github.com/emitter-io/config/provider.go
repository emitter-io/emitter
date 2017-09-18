package config

import (
	"errors"
	"os"
	"strings"
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
		return errors.New("Unable to configure Vault provider")
	}

	p.client = NewVaultClient(c.Vault().Address)
	p.app = c.Vault().Application

	// Authenticate the provider
	return p.client.Authenticate(p.app, p.user)
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
