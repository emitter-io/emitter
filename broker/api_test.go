package broker

import (
	"encoding/json"
	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/network/address"
	"github.com/emitter-io/emitter/security"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSingleContractProvider_Get(t *testing.T) {

	cfg, err := config.ReadOrCreate("emitter.conf", security.NewEnvironmentProvider(), security.NewVaultProvider(address.Fingerprint()))
	if err != nil {
		panic("Unable to parse configuration, due to " + err.Error())
	}

	// Setup the new service
	svc, err := NewService(cfg)
	if err != nil {
		panic(err.Error())
	}

	svc.ContractProvider = security.NewSingleContractProvider(svc.License)
	channel := security.ParseChannel([]byte("emitter/keygen/"))
	message := keyGenMessage{
		Key:     "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI",
		Channel: "test",
		Type:    "rw",
	}
	payload, _ := json.Marshal(&message)

	err = ProcessKeyGen(svc, channel, payload)
	assert.Nil(t, err)
	//assert.Nil(t, contractByWrongID)
}
