package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	cfg "github.com/emitter-io/config"
	"github.com/emitter-io/emitter/broker"
	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/security"
)

func main() {
	// Get the directory of the process
	exe, err := os.Executable()
	if err != nil {
		panic(err.Error())
	}

	// Process command-line arguments
	logging.LogTarget("service", "configured vault username", config.VaultUser)
	argConfig := flag.String("config", filepath.Join(filepath.Dir(exe), "emitter.conf"), "The configuration file to use for the broker.")
	argHelp := flag.Bool("help", false, "Shows the help and usage instead of running the broker.")
	flag.Parse()
	if *argHelp {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse the configuration
	c, err := cfg.ReadOrCreate("emitter", *argConfig, config.NewDefault,
		cfg.NewEnvironmentProvider(),
		cfg.NewVaultProvider(config.VaultUser))
	if err != nil {
		panic("Unable to parse configuration, due to " + err.Error())
	}

	// Generate a new license if none was provided
	cfg := c.(*config.Config)
	if cfg.License == "" {
		license, secret := security.NewLicenseAndMaster()
		logging.LogAction("service", "unable to find a license, make sure 'license' "+
			"value is set in the config file or EMITTER_LICENSE environment variable")
		logging.LogAction("service", fmt.Sprintf("generated new license: %v", license))
		logging.LogAction("service", fmt.Sprintf("generated new secret key: %v", secret))
		os.Exit(0)
	}

	// Setup the new service
	svc, err := broker.NewService(cfg)
	if err != nil {
		panic(err.Error())
	}

	// Listen and serve
	svc.Listen()
}
