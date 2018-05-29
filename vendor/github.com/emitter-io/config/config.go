package config

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// SecretStore represents a contract for a store capable of resolving secrets.
type SecretStore interface {
	Configure(config Config) error
	GetSecret(secretName string) (string, bool)
}

// CertificateStore represents a secret store which can be used for certificates.
type CertificateStore interface {
	autocert.Cache
}

// Provider represents a configurable provider.
type Provider interface {
	Name() string
	Configure(config map[string]interface{}) error
}

// Config represents a configuration interface.
type Config interface {
	Vault() *VaultConfig
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

	// Default to disk cache
	if certCache == nil {
		certCache = autocert.DirCache("certs")
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

// VaultConfig represents Vault configuration.
type VaultConfig struct {
	Address     string `json:"address"` // The vault address to use.
	Application string `json:"app"`     // The vault application ID to use.
}

// NewClient creates a new vault client for the configuration.
func (c *VaultConfig) NewClient(user string) (client *VaultClient, err error) {
	if c.Address == "" || c.Application == "" {
		return nil, errors.New("unable to configure Vault provider")
	}

	client = NewVaultClient(c.Address)
	err = client.Authenticate(c.Application, user)
	return
}

// ProviderConfig represents provider configuration.
type ProviderConfig struct {

	// The storage provider, this can either be specific builtin or the plugin path (file or
	// url) if the plugin is specified, it must contain a constructor function named 'New'
	// which returns an interface{}.
	Provider string `json:"provider"`

	// The configuration for a provider. This specifies various parameters to provide to the
	// specific provider during the Configure() call.
	Config map[string]interface{} `json:"config,omitempty"`
}

// LoadOrPanic loads a provider from the configuration and uses one or several builtins
// provided. If the provider is not found, it panics.
func (c *ProviderConfig) LoadOrPanic(builtins ...Provider) Provider {
	provider, err := c.Load(builtins...)
	if err != nil {
		panic(err)
	}

	return provider
}

// Load loads a provider from the configuration and uses one or several builtins provided.
func (c *ProviderConfig) Load(builtins ...Provider) (Provider, error) {
	for _, builtin := range builtins {
		if strings.ToLower(builtin.Name()) == strings.ToLower(c.Provider) {
			if err := builtin.Configure(c.Config); err != nil {
				return nil, errors.New("The provider '" + c.Provider + "' could not be loaded. " + err.Error())
			}

			return builtin, nil
		}
	}

	// Attempt to load a plugin provider
	p, err := plugin.Open(resolvePath(c.Provider))
	if err != nil {
		return nil, errors.New("The provider plugin '" + c.Provider + "' could not be opened. " + err.Error())
	}

	// Get the symbol
	sym, err := p.Lookup("New")
	if err != nil {
		return nil, errors.New("The provider '" + c.Provider + "' does not contain 'func New() interface{}' symbol")
	}

	// Resolve the
	pFactory, validFunc := sym.(*func() interface{})
	if !validFunc {
		return nil, errors.New("The provider '" + c.Provider + "' does not contain 'func New() interface{}' symbol")
	}

	// Construct the provider
	provider, validProv := ((*pFactory)()).(Provider)
	if !validProv {
		return nil, errors.New("The provider '" + c.Provider + "' does not implement 'Provider'")
	}

	// Configure the provider
	err = provider.Configure(c.Config)
	if err != nil {
		return nil, errors.New("The provider '" + c.Provider + "' could not be configured")
	}

	// Succesfully opened and configured a provider
	return provider, nil
}

// LoadProvider loads a provider from the configuration or panics if the configuration is
// specified, but the provider was not found or not able to configure. This uses the first
// provider as a default value.
func LoadProvider(config *ProviderConfig, providers ...Provider) Provider {
	if config == nil || config.Provider == "" {
		config = &ProviderConfig{
			Provider: providers[0].Name(),
		}
	}

	// Load the provider according to the configuration
	return config.LoadOrPanic(providers...)
}

// Write writes the configuration to a specific writer, in JSON format.
func write(config interface{}, output io.Writer) (int, error) {
	var formatted bytes.Buffer
	body, err := json.Marshal(config)
	if err != nil {
		return 0, err
	}

	if err := json.Indent(&formatted, body, "", "\t"); err != nil {
		return 0, err
	}

	return output.Write(formatted.Bytes())
}

// createDefault writes the default configuration to disk.
func createDefault(path string, newDefault func() Config) (Config, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	c := newDefault()
	if _, err := write(c, f); err != nil {
		return nil, err
	}
	if err := f.Sync(); err != nil {
		return nil, err
	}
	return c, nil
}

// ReadOrCreate reads or creates the configuration object.
func ReadOrCreate(prefix string, path string, newDefault func() Config, stores ...SecretStore) (cfg Config, err error) {
	cfg = newDefault()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create a configuration and write it to a file
		if cfg, err = createDefault(path, newDefault); err != nil {
			return nil, err
		}
	} else {
		// Read the config from file
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		// Unmarshal the configuration
		if err := json.Unmarshal(b, cfg); err != nil {
			return nil, err
		}
	}

	// Apply all the store overrides, in order
	for _, store := range stores {
		if err := store.Configure(cfg); err == nil {
			declassify(cfg, prefix, store)
		}
	}

	return cfg, nil
}

// Declassify traverses the configuration and resolves secrets.
func declassify(config interface{}, prefix string, provider SecretStore) {
	original := reflect.ValueOf(config)
	declassifyRecursive(prefix, provider, original)
}

// DeclassifyRecursive traverses the configuration and resolves secrets.
func declassifyRecursive(prefix string, provider SecretStore, value reflect.Value) {
	switch value.Kind() {
	case reflect.Ptr:
		pValue := value.Elem()
		if !pValue.IsValid() {
			// Create a new struct and set the value
			pValue = reflect.New(value.Type().Elem())
			value.Set(pValue)
		}

		declassifyRecursive(prefix, provider, pValue)

	// If it is a struct we translate each field
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			name := getFieldName(value.Type().Field(i))
			if name != "" { // If there's no JSON tag, ignore it
				declassifyRecursive(prefix+"/"+name, provider, value.Field(i))
			}
		}

	// This is a integer, we need to fetch the secret
	case reflect.Int:
		if v, ok := provider.GetSecret(prefix); ok {
			if iv, err := strconv.ParseInt(v, 10, 64); err == nil {
				value.SetInt(iv)
			}
		}

	// This is a string, we need to fetch the secret
	case reflect.String:
		if v, ok := provider.GetSecret(prefix); ok {
			value.SetString(v)
		}

	// This is a map, unmarshal and set
	case reflect.Map:
		if v, ok := provider.GetSecret(prefix); ok {
			var out map[string]interface{}
			if err := json.Unmarshal([]byte(v), &out); err == nil {
				value.Set(reflect.ValueOf(out))
			}
		}
	}
}

func getFieldName(f reflect.StructField) string {
	return strings.Replace(string(f.Tag.Get("json")), ",omitempty", "", -1)
}

func resolvePath(path string) string {

	// If it's an url, download the file
	if strings.HasPrefix(path, "http") {
		f, err := httpFile(path)
		if err != nil {
			panic(err)
		}

		// Get the downloaded file path
		path = f.Name()
	}

	// Make sure the path is absolute
	path, _ = filepath.Abs(path)
	return path
}
