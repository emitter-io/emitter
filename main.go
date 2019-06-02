package main

import (
	"context"
	"fmt"
	"os"

	"github.com/emitter-io/config/dynamo"
	"github.com/emitter-io/config/vault"
	"github.com/emitter-io/emitter/internal/broker"
	"github.com/emitter-io/emitter/internal/command/load"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/security/license"
	"github.com/jawher/mow.cli"
)

//go:generate go run internal/broker/generate/assets_gen.go

func main() {
	app := cli.App("emitter", "Emitter: the high performance, distributed and low latency publish-subscribe platform.")
	app.Spec = "[ -c=<configuration path> ] "
	confPath := app.StringOpt("c config", "emitter.conf", "The configuration file to use for the broker.")
	app.Action = func() { run(confPath) }

	// Register sub-commands
	app.Command("load", "Runs the load testing client for emitter.", load.Run)
	app.Run(os.Args)

}

func run(conf *string) {
	cfg := config.New(*conf, dynamo.NewProvider(), vault.NewProvider(config.VaultUser))

	// Generate a new license if none was provided
	if cfg.License == "" {
		license, secret := license.New()
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
