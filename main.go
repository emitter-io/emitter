package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/emitter-io/config/dynamo"
	"github.com/emitter-io/config/vault"
	"github.com/emitter-io/emitter/internal/broker"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/security"
)

func main() {
	// Get the directory of the process
	exe, err := os.Executable()
	if err != nil {
		panic(err.Error())
	}

	// Process command-line arguments
	argConfig := flag.String("config", filepath.Join(filepath.Dir(exe), "emitter.conf"), "The configuration file to use for the broker.")
	argHelp := flag.Bool("help", false, "Shows the help and usage instead of running the broker.")
	flag.Parse()
	if *argHelp {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read the configuration
	cfg := config.New(*argConfig,
		dynamo.NewProvider(),
		vault.NewProvider(config.VaultUser),
	)

	// Generate a new license if none was provided
	if cfg.License == "" {
		license, secret := security.NewLicenseAndMaster()
		logging.LogAction("service", "unable to find a license, make sure 'license' "+
			"value is set in the config file or EMITTER_LICENSE environment variable")
		logging.LogAction("service", fmt.Sprintf("generated new license: %v", license))
		logging.LogAction("service", fmt.Sprintf("generated new secret key: %v", secret))
		os.Exit(0)
	}

	// Start new service
	svc, err := broker.NewService(context.Background(), cfg)
	if err != nil {
		panic(err.Error())
	}

	// Listen and serve
	svc.Listen()
}
