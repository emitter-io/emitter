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
	cfg.License = testLicense
	cfg.Cluster = nil
	svc, _ := NewService(cfg)
	svc.ContractProvider = security.NewSingleContractProvider(svc.License)

	channel := security.ParseChannel([]byte("emitter/keygen/"))
	message := keyGenMessage{
		Key:     "kBCZch5re3Ue-kpG1Aa8Vo7BYvXZ3UwR",
		Channel: "a",
		Type:    "rw",
	}
	payload, _ := json.Marshal(&message)

	s, err := KeyGen(svc, channel, payload, true, 1)
	assert.Nil(t, err)
	assert.Equal(t, "0Nq8SWbL8qoOKEDqh_ebBepug6cLLlWO", s)
}
