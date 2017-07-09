package main

import (
	"math/rand"
	"os"
	"time"

	"bitbucket.org/emitter/emitter/config"
	"bitbucket.org/emitter/emitter/emitter"
	"bitbucket.org/emitter/emitter/logging"
	"bitbucket.org/emitter/emitter/network/address"
	"bitbucket.org/emitter/emitter/security"
	"bitbucket.org/emitter/emitter/utils"
)

func main() {
	logging.SetWriter(os.Stdout, true)

	// Parse the configuration
	cfg, err := config.ReadOrCreate("emitter.conf", security.NewEnvironmentProvider(), security.NewVaultProvider(address.Fingerprint()))
	if err != nil {
		panic("Unable to parse configuration, due to " + err.Error())
	}

	// Setup the new service
	svc, err := emitter.NewService(cfg)
	if err != nil {
		panic(err.Error())
	}

	// Flush the log
	utils.Repeat(func() {
		if err := logging.Flush(); err != nil {
			println("Unable to flush logger, due to " + err.Error())
		}
	}, 100*time.Millisecond, svc.Closing)

	// Initialize the rand function for key generation.
	rand.Seed(time.Now().UTC().UnixNano())

	// Listen and serve
	svc.Listen()
}
