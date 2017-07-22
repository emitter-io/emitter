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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/emitter-io/emitter/network/http"
)

// VaultClient represents a lightweight vault client.
type vaultClient struct {
	address string // The vault address.
	token   string // The vault token provided by auth.
}

// NewVaultClient creates a new vault client.
func newVaultClient(address string) *vaultClient {
	if ip := net.ParseIP(address); ip != nil {
		address = fmt.Sprintf("http://%v:8200", ip.String())
	}

	return &vaultClient{
		address: address,
	}
}

// IsAuthenticated checks whether we are authenticated or not.
func (c *vaultClient) IsAuthenticated() bool {
	return c.token != ""
}

// Authenticate performs vault authentication.
func (c *vaultClient) Authenticate(app string, user string) error {
	output, err := c.post("/auth/app-id/login", &vaultAuthRequest{
		App:  app,
		User: user,
	})

	// Unable to authenticate with Vault
	if err == nil && output.Auth != nil {
		c.token = output.Auth.ClientToken
		return nil
	}

	return errors.New("Unable to perform vault authentication for user " + user)
}

// ReadSecret reads a secret from the vault.
func (c *vaultClient) ReadSecret(secretName string) (string, error) {
	output, err := c.get("/secret/" + secretName)
	if err != nil {
		return "", err
	}

	if secret, ok := output.Data["value"]; ok {
		value := secret.(string)
		if value != "" {
			return value, nil
		}
	}

	return "", errors.New("Unable to find or parse secret " + secretName)
}

// ReadSecret reads a secret from the vault.
func (c *vaultClient) ReadCredentials(credentialsName string) (*AwsCredentials, error) {
	output, err := c.get("/aws/sts/" + credentialsName)
	if err != nil {
		return nil, err
	}

	if s, err := json.Marshal(output.Data); err == nil {
		creds := new(AwsCredentials)
		if err := json.Unmarshal(s, creds); err == nil && creds.AccessKey != "" {
			creds.Duration = time.Duration(output.LeaseDuration) * time.Second
			creds.Expires = time.Now().UTC().Add(creds.Duration)
			return creds, nil
		}
	}

	return nil, errors.New("Unable to find or parse credentials " + credentialsName)
}

// Get issues an HTTP GET to a vault server.
func (c *vaultClient) get(url string) (output *vaultSecret, err error) {
	var headers []http.HeaderValue
	if c.IsAuthenticated() {
		headers = append(headers, http.NewHeader("X-Vault-Token", c.token))
	}

	// Issue the HTTP Get
	output = new(vaultSecret)
	err = http.Get(c.address+"/v1"+url, output, headers...)
	return
}

// Post issues an HTTP POST to a vault server.
func (c *vaultClient) post(url string, body interface{}) (output *vaultSecret, err error) {

	// Note: The HTTP post is used for authentication only right now, hence no need
	// to add the X-Vault-Token.
	var headers []http.HeaderValue
	output = new(vaultSecret)
	err = http.Post(c.address+"/v1"+url, body, output, headers...)
	return
}

// vaultAuthRequest is the structure representing a request for authentication.
type vaultAuthRequest struct {
	App  string `json:"app_id"`
	User string `json:"user_id"`
}

// vaultSecret is the structure returned for every secret within Vault.
type vaultSecret struct {
	RequestID     string                 `json:"request_id"`
	LeaseID       string                 `json:"lease_id"`
	LeaseDuration int                    `json:"lease_duration"`
	Renewable     bool                   `json:"renewable"`
	Data          map[string]interface{} `json:"data"`
	Warnings      []string               `json:"warnings"`
	Auth          *vaultSecretAuth       `json:"auth,omitempty"`
	WrapInfo      *vaultSecretWrapInfo   `json:"wrap_info,omitempty"`
}

// vaultSecretWrapInfo contains wrapping information if we have it.
type vaultSecretWrapInfo struct {
	Token           string    `json:"token"`
	TTL             int       `json:"ttl"`
	CreationTime    time.Time `json:"creation_time"`
	WrappedAccessor string    `json:"wrapped_accessor"`
}

// SecretAuth is the structure containing auth information if we have it.
type vaultSecretAuth struct {
	ClientToken   string            `json:"client_token"`
	Accessor      string            `json:"accessor"`
	Policies      []string          `json:"policies"`
	Metadata      map[string]string `json:"metadata"`
	LeaseDuration int               `json:"lease_duration"`
	Renewable     bool              `json:"renewable"`
}

// AwsCredentials represents Amazon Web Services credentials.
type AwsCredentials struct {
	AccessKey string        `json:"access_key"`     // The access key.
	SecretKey string        `json:"secret_key"`     // The secret key.
	Token     string        `json:"security_token"` // The token.
	Duration  time.Duration `json:"-"`              // The duration of the credentials.
	Expires   time.Time     `json:"-"`              // The expiration date of the credentials.
}
