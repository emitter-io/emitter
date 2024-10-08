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

package keygen

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/service/keygen"
	cli "github.com/jawher/mow.cli"
)

// Generate a new channel key.
func NewKey(cmd *cli.Cmd) {
	cmd.Spec = "MASTERKEY CHANNEL ACCESS [ -t=<ttl>] [ -h=<host> ]"
	var (
		masterkey = cmd.StringArg("MASTERKEY", "", "Specifies the master key for generating channel keys.")
		channel   = cmd.StringArg("CHANNEL", "", "Specifies the name channel for which to generate key.")
		access    = cmd.StringArg("ACCESS", "", "Specifies the access rights for the channel (rwslpex).")
		ttl       = cmd.IntOpt("t ttl", 0, "Specifies the time to live for the key in seconds. By default, the key will never expire.")
		host      = cmd.StringOpt("h host", "127.0.0.1:8080", "Specifies the broker host name and port. This must follow the <ip:port> format.")
	)

	cmd.Action = func() {
		// Create a channel to receive the response.
		respChan := make(chan []byte)

		// Create a new MQTT client.
		opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", *host))
		opts.SetDefaultPublishHandler(func(client mqtt.Client, m mqtt.Message) {
			respChan <- m.Payload()
		})
		client := mqtt.NewClient(opts)

		// Connect to the MQTT broker.
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(fmt.Errorf("failed to connect: %v", token.Error()))
		}

		// Publish the request.
		request := keygen.Request{
			Key:     *masterkey,
			Channel: *channel,
			Type:    *access,
			TTL:     int32(*ttl),
		}
		payload, err := json.Marshal(request)
		if err != nil {
			logging.LogError("keygen", "marshaling the request", err)
			return
		}
		if token := client.Publish("emitter/keygen/", 1, false, payload); token.Wait() && token.Error() != nil {
			logging.LogError("keygen", "publishing the request", token.Error())
			return
		}

		// Wait for the response and check whether it's an error.
		response := <-respChan
		var errResponse errors.Error
		if err := json.Unmarshal(response, &errResponse); err == nil && errResponse.Error() != "" {
			logging.LogError("keygen", "generating the key", fmt.Errorf("error: %s", errResponse.Error()))
			return
		}

		// Parse the response and print the key.
		keygen := keygen.Response{}
		if err := json.Unmarshal(response, &keygen); err != nil {
			logging.LogError("keygen", "parsing the response", err)
			return
		}
		fmt.Println(keygen.Key)
		client.Disconnect(250)
	}
}
