// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package config

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// CertCacher represents a contract which allows for retrieval of certificate cache
type CertCacher interface {
	Name() string
	GetCache() (autocert.Cache, bool)
}

// DirCache implements Cache using a directory on the local filesystem.
// If the directory does not exist, it will be created with 0700 permissions.
type dirCache struct{}

func (c *dirCache) Name() string {
	return "dircache"
}

func (c *dirCache) GetCache() (autocert.Cache, bool) {
	return autocert.DirCache("certs"), true
}

// TLS returns a TLS configuration which can be applied or validated. This requires a set of valid
// stores in the first place, so for this to work it needs to be called once the stores are configured
// (aka Configure() method was called). This should work well if called at some point after calling
// config.ReadOrCreate().
func TLS(cfg *TLSConfig, stores ...CertCacher) (*tls.Config, http.Handler, CertCacher) {
	stores = append(stores, new(dirCache)) // Fallback to DirCache

	// Go through all of the certificate stores
	for _, store := range stores {
		if cache, valid := store.GetCache(); valid {
			if tls, val, err := cfg.Load(cache); err == nil {
				return tls, val, store
			}
		}
	}

	return nil, nil, nil
}

// TLSConfig represents TLS listener configuration.
type TLSConfig struct {
	ListenAddr  string `json:"listen"`                // The address to listen on.
	Host        string `json:"host"`                  // The hostname to whitelist.
	Email       string `json:"email,omitempty"`       // The email address for autocert.
	Certificate string `json:"certificate,omitempty"` // The certificate request.
	PrivateKey  string `json:"private,omitempty"`     // The private key for the certificate.
}

// Load loads the certificates from the cache or the configuration.
func (c *TLSConfig) Load(certCache autocert.Cache) (*tls.Config, http.Handler, error) {
	if c.Certificate != "" {
		return c.loadFromLocal(certCache)
	}

	return c.loadFromAutocert(certCache)
}

// loadFromLocal loads TLS configuration from pre-existing certificate
func (c *TLSConfig) loadFromLocal(certCache autocert.Cache) (*tls.Config, http.Handler, error) {
	if c.PrivateKey == "" {
		return &tls.Config{}, nil, errors.New("No certificate or private key configured")
	}

	// If the certificate provided is in plain text, write to file so we can read it.
	if strings.HasPrefix(c.Certificate, "---") {
		if err := ioutil.WriteFile("broker.crt", []byte(c.Certificate), os.ModePerm); err == nil {
			c.Certificate = "broker.crt"
		}
	}

	// If the private key provided is in plain text, write to file so we can read it.
	if strings.HasPrefix(c.PrivateKey, "---") {
		if err := ioutil.WriteFile("broker.key", []byte(c.PrivateKey), os.ModePerm); err == nil {
			c.PrivateKey = "broker.key"
		}
	}

	// Make sure the paths are absolute, otherwise we won't be able to read the files.
	c.Certificate = resolvePath(c.Certificate)
	c.PrivateKey = resolvePath(c.PrivateKey)

	// Load the certificate from the cert/key files.
	cer, err := tls.LoadX509KeyPair(c.Certificate, c.PrivateKey)
	return &tls.Config{
		Certificates: []tls.Certificate{cer},
	}, nil, err
}

// loadFromAutocert loads TLS configuration from Letsencrypt
func (c *TLSConfig) loadFromAutocert(certCache autocert.Cache) (*tls.Config, http.Handler, error) {
	if c.Host == "" {
		return nil, nil, errors.New("unable to request a certificate, no host name configured")
	}

	// Create an auto-cert manager
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(c.Host),
		Email:      c.Email,
		Cache:      certCache,
	}

	return &tls.Config{
		GetCertificate: certManager.GetCertificate,
	}, certManager.HTTPHandler(nil), nil
}
