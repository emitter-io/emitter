package config

import "errors"

func (c *ProviderConfig) LoadPlugin() (Provider, error) {
	return nil, errors.New("The provider plugin '" + c.Provider + "' could not be opened. Plugin not supported on darwin.")
}
