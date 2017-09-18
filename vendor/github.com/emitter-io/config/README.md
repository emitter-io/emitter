# config

This repository contains the configuration parsing and management package designed for Emitter (https://emitter.io) service and related services. The configuration contains a flexible and multi-level secret overrides with `Environment Variable` and `Hashicorp Vault` providers implemented out of the box.


## Installation

If you want to use this package, you can simply `go get` it as shown below.

```
go get github.com/emitter-io/config
```

## Usage

```
import (
	cfg "github.com/emitter-io/config"
)

// NewDefault creates a default configuration.
func NewDefault() cfg.Config {
	return &Config{
		ListenAddr: ":8080",
		Cluster: &ClusterConfig{
			ListenAddr:    ":4000",
			AdvertiseAddr: "public:4000",
			Passphrase:    "emitter-io",
		},
	}
}

// Config represents main configuration.
type Config struct {
	ListenAddr string              `json:"listen"`             // The API port used for TCP & Websocket communication.'
	License    string              `json:"license"`            // The port used for gossip.'
	TLS        *cfg.TLSConfig      `json:"tls,omitempty"`      // The API port used for Secure TCP & Websocket communication.'
	Secrets    *cfg.VaultConfig    `json:"vault,omitempty"`    // The configuration for the Hashicorp Vault.
	Storage    *cfg.ProviderConfig `json:"storage,omitempty"`  // The configuration for the storage provider.
	Contract   *cfg.ProviderConfig `json:"contract,omitempty"` // The configuration for the contract provider.
}

// Vault returns a vault configuration.
func (c *Config) Vault() *cfg.VaultConfig {
	return c.Secrets
}

func main() {
	// Parse the configuration
	cfg, err := config.ReadOrCreate("emitter", "config.json", NewDefault, config.NewEnvironmentProvider(), config.NewVaultProvider("app"))
	if err != nil {
		panic("Unable to parse configuration, due to " + err.Error())
	}

    ...
}

```
