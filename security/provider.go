/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/

package security

import (
	"errors"
	"os"
	"strings"

	"github.com/emitter-io/emitter/config"
)

// Provider represents a contract for a security provider.
type Provider interface {
	config.SecretStore
}

// ------------------------------------------------------------------------------------

// EnvironmentProvider represents a security provider which uses environment variables to store secrets.
type EnvironmentProvider struct {
}

// NewEnvironmentProvider creates a new environment security provider.
func NewEnvironmentProvider() Provider {
	return new(EnvironmentProvider)
}

// Configure configures the security provider.
func (p *EnvironmentProvider) Configure(c *config.Config) error {
	return nil
}

// GetSecret retrieves a secret from the provider
func (p *EnvironmentProvider) GetSecret(secretName string) (string, bool) {
	name := strings.ToUpper(strings.Replace(secretName, "/", "_", -1))
	return os.LookupEnv(name)
}

// ------------------------------------------------------------------------------------

// VaultProvider represents a security provider which uses hashicorp vault to store secrets.
type VaultProvider struct {
	client *vaultClient // The vault client.
	app    string       // The application ID to use for authentication.
	user   string       // The user ID to use for authentication.
	auth   bool         // Whether the provider is authenticated or not.
}

// NewVaultProvider creates a new environment security provider.
func NewVaultProvider(user string) *VaultProvider {
	return &VaultProvider{user: user}
}

// Configure configures the security provider.
func (p *VaultProvider) Configure(c *config.Config) error {
	if !c.HasVault() {
		p.client = newVaultClient(c.Vault.Address)
		p.app = c.Vault.Application

		// Authenticate the provider
		if err := p.client.Authenticate(p.app, p.user); err != nil {
			return err
		}
		return nil
	}
	return errors.New("Unable to configure Vault provider")
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
