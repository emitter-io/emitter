/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/

package main

import (
	"context"
	"os"

	"github.com/emitter-io/config/dynamo"
	"github.com/emitter-io/config/vault"
	"github.com/emitter-io/emitter/internal/broker"
	"github.com/emitter-io/emitter/internal/command/license"
	"github.com/emitter-io/emitter/internal/command/load"
	"github.com/emitter-io/emitter/internal/command/version"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/provider/logging"
	cli "github.com/jawher/mow.cli"
)

//go:generate go run internal/broker/generate/assets_gen.go

func main() {
	app := cli.App("emitter", "Runs the Emitter broker.")
	app.Spec = "[ -c=<configuration path> ] "
	confPath := app.StringOpt("c config", "emitter.conf", "Specifies the configuration path (file) to use for the broker.")
	app.Action = func() { listen(app, confPath) }

	// Register sub-commands
	app.Command("version", "Prints the version of the executable.", version.Print)
	app.Command("load", "Runs the load testing client for emitter.", load.Run)
	app.Command("license", "Manipulates licenses and secret keys.", func(cmd *cli.Cmd) {
		cmd.Command("new", "Generates a new license and secret key pair.", license.New)
		// TODO: add more sub-commands for license
	})

	app.Run(os.Args)
}

// Listen starts the service.
func listen(app *cli.Cli, conf *string) {

	// Generate a new license if none was provided
	cfg := config.New(*conf, dynamo.NewProvider(), vault.NewProvider(config.VaultUser))
	if cfg.License == "" {
		logging.LogAction("service", "unable to find a license, make sure 'license' "+
			"value is set in the config file or EMITTER_LICENSE environment variable")
		app.Run([]string{"emitter", "license", "new"})
		return
	}

	// Start new service
	svc, err := broker.NewService(context.Background(), cfg)
	if err != nil {
		logging.LogError("service", "startup", err)
		return
	}

	// Listen and serve
	svc.Listen()
}
