package broker

import (
	"encoding/json"
	"testing"

	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/security"
	"github.com/stretchr/testify/assert"
)

func TestProcessKeygen(t *testing.T) {

	cfg := config.NewDefault()
	cfg.License = "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI"
	cfg.Cluster = nil
	svc, _ := NewService(cfg)
	svc.ContractProvider = security.NewSingleContractProvider(svc.License)

	channel := security.ParseChannel([]byte("emitter/keygen/"))
	message := keyGenMessage{
		Key:     "kBCZch5re3Ue-kpG1Aa8Vo7BYvXZ3UwR",
		Channel: "test",
		Type:    "rw",
	}
	payload, _ := json.Marshal(&message)

	err := ProcessKeyGen(svc, channel, payload)
	assert.Nil(t, err)
}
