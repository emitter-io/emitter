package main

import (
	"flag"
	"os"
	"time"

	"github.com/emitter-io/emitter/broker"
	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/logging"
	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/emitter/security"
	"github.com/emitter-io/emitter/utils"
)

func main() {
	logging.SetWriter(os.Stdout, true)

	// Process command-line arguments
	argConfig := flag.String("config", "emitter.conf", "The configuration file to use for the broker.")
	argHelp := flag.Bool("help", false, "Shows the help and usage instead of running the broker.")
	flag.Parse()
	if *argHelp {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse the configuration
	// TODO: emitter.conf should come from command line args
	cfg, err := config.ReadOrCreate(*argConfig, security.NewEnvironmentProvider(), security.NewVaultProvider(address.Fingerprint()))
	if err != nil {
		panic("Unable to parse configuration, due to " + err.Error())
	}

	// Setup the new service
	svc, err := broker.NewService(cfg)
	if err != nil {
		panic(err.Error())
	}
	svc.Contracts = security.NewSingleContractProvider(svc.License)

	// Flush the log
	utils.Repeat(func() {
		if err := logging.Flush(); err != nil {
			println("Unable to flush logger, due to " + err.Error())
		}
	}, 100*time.Millisecond, svc.Closing)

	// Listen and serve
	svc.Listen()
}
