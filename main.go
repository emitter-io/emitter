package main

import (
	"flag"
	"os"

	cfg "github.com/emitter-io/config"
	"github.com/emitter-io/emitter/broker"
	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/network/address"
)

func main() {
	// Process command-line arguments
	argConfig := flag.String("config", "emitter.conf", "The configuration file to use for the broker.")
	argHelp := flag.Bool("help", false, "Shows the help and usage instead of running the broker.")
	flag.Parse()
	if *argHelp {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse the configuration
	cfg, err := cfg.ReadOrCreate("emitter", *argConfig, config.NewDefault, cfg.NewEnvironmentProvider(), cfg.NewVaultProvider(address.Hardware().Hex()))
	if err != nil {
		panic("Unable to parse configuration, due to " + err.Error())
	}

	// Setup the new service
	svc, err := broker.NewService(cfg.(*config.Config))
	if err != nil {
		panic(err.Error())
	}

	//secret, _ := svc.License.NewMasterKey(1)
	//c, _ := svc.Cipher.GenerateKey(secret, "cluster", security.AllowRead, time.Unix(0, 0))
	//println(c)

	// Listen and serve
	svc.Listen()

}
