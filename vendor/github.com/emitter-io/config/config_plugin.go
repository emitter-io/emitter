// +build !darwin

package config

import (
	"errors"
	"plugin"
)

//
func (c *ProviderConfig) LoadPlugin() (Provider, error) {
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
