package config

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// VaultClient represents a lightweight vault client.
type VaultClient struct {
	address string // The vault address.
	token   string // The vault token provided by auth.
}

// NewVaultClient creates a new vault client.
func NewVaultClient(address string) *VaultClient {
	if ip := net.ParseIP(address); ip != nil {
		address = fmt.Sprintf("http://%v:8200", ip.String())
	}

	return &VaultClient{
		address: address,
	}
}

// IsAuthenticated checks whether we are authenticated or not.
func (c *VaultClient) IsAuthenticated() bool {
	return c.token != ""
}

// Authenticate performs vault authentication.
func (c *VaultClient) Authenticate(app string, user string) error {
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
func (c *VaultClient) ReadSecret(secretName string) (string, error) {
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

// WriteSecret writes a secret to the vault.
func (c *VaultClient) WriteSecret(secretName string, value string) error {
	_, err := c.post("/secret/"+secretName, map[string]string{
		"value": value,
	})
	return err
}

// Get issues an HTTP GET to a vault server.
func (c *VaultClient) get(url string) (output *vaultSecret, err error) {
	var headers []httpHeader
	if c.IsAuthenticated() {
		headers = append(headers, newHttpHeader("X-Vault-Token", c.token))
	}

	// Issue the HTTP Get
	output = new(vaultSecret)
	err = httpGet(c.address+"/v1"+url, output, headers...)
	return
}

// Post issues an HTTP POST to a vault server.
func (c *VaultClient) post(url string, body interface{}) (output *vaultSecret, err error) {
	var headers []httpHeader
	if c.IsAuthenticated() {
		headers = append(headers, newHttpHeader("X-Vault-Token", c.token))
	}

	// Issue the HTTP Post
	output = new(vaultSecret)
	err = httpPost(c.address+"/v1"+url, body, output, headers...)
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
