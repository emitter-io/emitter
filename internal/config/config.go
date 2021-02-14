/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
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

package config

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"

	"github.com/emitter-io/address"
	cfg "github.com/emitter-io/config"
	"github.com/emitter-io/emitter/internal/provider/logging"
)

// Constants used throughout the service.
const (
	ChannelSeparator = '/'   // The separator character.
	maxMessageSize   = 65536 // Default Maximum message size allowed from/to the peer.
)

// VaultUser is the vault user to use for authentication
var VaultUser = toUsername(address.GetExternalOrDefault(address.Loopback))

// Type alias for a raw config
type secretStoreConfig = map[string]interface{}

// toUsername converts an ip address to a username for Vault.
func toUsername(a net.IPAddr) string {
	return strings.Replace(
		strings.Replace(a.IP.String(), ".", "-", -1),
		":", "-", -1)
}

// NewDefault creates a default configuration.
func NewDefault() cfg.Config {
	return &Config{
		ListenAddr: ":8080",
		TLS: &cfg.TLSConfig{
			ListenAddr: ":443",
		},
		Cluster: &ClusterConfig{
			ListenAddr:    ":4000",
			AdvertiseAddr: "external:4000",
		},
		Storage: &cfg.ProviderConfig{
			Provider: "inmemory",
		},
	}
}

// New reads or creates a configuration.
func New(filename string, stores ...cfg.SecretStore) *Config {
	readers := []cfg.SecretReader{cfg.NewEnvironmentProvider()}
	caches := []cfg.CertCacher{}
	for _, store := range stores {
		readers = append(readers, store)
		caches = append(caches, store)
	}

	c, err := cfg.ReadOrCreate("emitter", filename, NewDefault, readers...)
	if err != nil {
		panic("Unable to parse configuration, due to " + err.Error())
	}

	conf := c.(*Config)
	conf.certCaches = caches
	return conf
}

// Config represents main configuration.
type Config struct {
	ListenAddr string              `json:"listen"`             // The API port used for TCP & Websocket communication.
	License    string              `json:"license"`            // The license file to use for the broker.
	Matcher    string              `json:"matcher,omitempty"`  // If "mqtt", then topic matching would follow MQTT specification.
	Debug      bool                `json:"debug,omitempty"`    // The debug mode flag.
	Limit      LimitConfig         `json:"limit,omitempty"`    // Configuration for various limits such as message size.
	TLS        *cfg.TLSConfig      `json:"tls,omitempty"`      // The API port used for Secure TCP & Websocket communication.
	Cluster    *ClusterConfig      `json:"cluster,omitempty"`  // The configuration for the clustering.
	Storage    *cfg.ProviderConfig `json:"storage,omitempty"`  // The configuration for the storage provider.
	Contract   *cfg.ProviderConfig `json:"contract,omitempty"` // The configuration for the contract provider.
	Metering   *cfg.ProviderConfig `json:"metering,omitempty"` // The configuration for the usage storage for metering.
	Logging    *cfg.ProviderConfig `json:"logging,omitempty"`  // The configuration for the logger.
	Monitor    *cfg.ProviderConfig `json:"monitor,omitempty"`  // The configuration for the monitoring storage.
	Vault      secretStoreConfig   `json:"vault,omitempty"`    // The configuration for the Hashicorp Vault Secret Store.
	Dynamo     secretStoreConfig   `json:"dynamodb,omitempty"` // The configuration for the AWS DynamoDB Secret Store.

	listenAddr *net.TCPAddr     // The listen address, parsed.
	certCaches []cfg.CertCacher // The certificate caches configured.
}

// MaxMessageBytes returns the configured max message size, must be smaller than 64K.
func (c *Config) MaxMessageBytes() int64 {
	if c.Limit.MessageSize <= 0 || c.Limit.MessageSize > maxMessageSize {
		return maxMessageSize
	}
	return int64(c.Limit.MessageSize)
}

// Addr returns the listen address configured.
func (c *Config) Addr() *net.TCPAddr {
	if c.listenAddr == nil {
		var err error
		if c.listenAddr, err = address.Parse(c.ListenAddr, 8080); err != nil {
			panic(err)
		}
	}
	return c.listenAddr
}

// Certificate returns TLS configuration.
func (c *Config) Certificate() (*tls.Config, http.Handler, bool) {
	if c.TLS == nil {
		return nil, nil, false
	}

	// Attempt to configure
	if tls, validator, cache := cfg.TLS(c.TLS, c.certCaches...); cache != nil {
		logging.LogAction("tls", "setting up certificates with "+cache.Name()+" cache")
		return tls, validator, true
	}

	logging.LogAction("tls", "unable to configure certificates, make sure a valid cache or certificate is configured")
	return nil, nil, false
}

// ClusterConfig represents the configuration for the cluster.
type ClusterConfig struct {

	// The name of this node. This must be unique in the cluster. If this is not set, Emitter
	// will set it to the external IP address of the running machine.
	NodeName string `json:"name,omitempty"`

	// The IP address and port that is used to bind the inter-node communication network. This
	// is used for the actual binding of the port.
	ListenAddr string `json:"listen"`

	// The address and port to advertise inter-node communication network. This is used for nat
	// traversal.
	AdvertiseAddr string `json:"advertise"`

	// The seed address (or a domain name) for cluster join.
	Seed string `json:"seed,omitempty"`

	// Passphrase is used to initialize the primary encryption key in a keyring. This key
	// is used for encrypting all the gossip messages (message-level encryption).
	Passphrase string `json:"passphrase,omitempty"`

	// Directory specifies the directory where the cluster state will be stored.
	Directory string `json:"dir,omitempty"`
}

// LimitConfig represents various limit configurations - such as message size.
type LimitConfig struct {

	// Maximum message size allowed from/to the client. Default if not specified is 64kB.
	MessageSize int `json:"messageSize,omitempty"`

	// The maximum messages per second allowed to be processed per client connection. This
	// effectively restricts the QpS for an individual connection.
	ReadRate int `json:"readRate,omitempty"`

	// The maximum socket write rate per connection. This does not limit QpS but instead
	// can be used to scale throughput. Defaults to 60.
	FlushRate int `json:"flushRate,omitempty"`
}

// LoadProvider loads a provider from the configuration or panics if the configuration is
// specified, but the provider was not found or not able to configure. This uses the first
// provider as a default value.
var LoadProvider = cfg.LoadProvider
