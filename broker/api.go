package broker

import (
	"crypto/rand"
	"encoding/json"
	"math"
	"math/big"
	"time"

	"github.com/emitter-io/emitter/security"
	"github.com/emitter-io/emitter/utils"
)

const (
	RequestKeygen   = 548658350
	RequestPresence = 3869262148
)

type keyGenMessage struct {
	Key     string `json:"key"`
	Channel string `json:"channel"`
	Type    string `json:"type"`
	TTL     int32  `json:"ttl"`
}

func (m *keyGenMessage) expires() time.Time {
	if m.TTL == 0 {
		return time.Time{}
	}
	return time.Now().Add(time.Second).UTC()
}

func (m *keyGenMessage) access() uint32 {
	required := security.AllowNone

	for i := 0; i < len(m.Type); i++ {
		switch c := m.Type[i]; c {
		case 'r':
			required |= security.AllowRead
		case 'w':
			required |= security.AllowWrite
		case 's':
			required |= security.AllowStore
		case 'l':
			required |= security.AllowLoad
		case 'p':
			required |= security.AllowPresence
		}
	}

	return required
}

// TryProcessAPIRequest attempts to generate the key and returns the result.
func TryProcessAPIRequest(c *Conn, channel *security.Channel, payload []byte) bool {
	defer func() {
		// Send the response, always
	}()

	/* The first version of Emitter established that any api request should starts with the word "emitter".
	 * It was implemented as a string comparison. This meant 7 comparisons (7 letters) for each message received,
	 * just to check whether the message was an api request.
	 * It was suggested that, for the sake of optimisation, an api request could be simply announced by the dummy channel '0'.
	 * However this meant breaking the code of already existing clients. Therefore, it was decided to just check whether the key
	 * was 7-letter long, which achieve mostly the same effect, since no valid key is supposed to be 7-byte long. */

	if len(channel.Key) != 7 {
		return false
	}

	switch channel.Query[1] {
	case RequestKeygen:
		ProcessKeyGen(c.service, channel, payload)
		return true
	case RequestPresence:
		return true
	default:
		return true
	}

}

// ProcessKeyGen processes a keygen request.
func ProcessKeyGen(s *Service, channel *security.Channel, payload []byte) error {
	// Deserialize the payload.
	message := keyGenMessage{}
	if err := json.Unmarshal(payload, &message); err != nil {
		return err
	}

	// Attempt to parse the key, this should be a master key
	masterKey, err := s.Cipher.DecryptKey([]byte(message.Key))
	if err != nil {
		return err
	}
	if !masterKey.IsMaster() || masterKey.IsExpired() {
		return ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract := s.ContractProvider.Get(masterKey.Contract())
	if contract == nil {
		return ErrNotFound
	}

	// Validate the contract
	if !contract.Validate(masterKey) {
		return ErrUnauthorized
	}

	// Generate the key
	key := security.Key(make([]byte, 24))

	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt16))
	key.SetSalt(uint16(n.Uint64()))
	key.SetMaster(masterKey.Master())
	key.SetContract(masterKey.Contract())
	key.SetSignature(masterKey.Signature())
	key.SetPermissions(message.access())
	key.SetTarget(utils.GetHash([]byte(message.Channel)))
	key.SetExpires(message.expires())

	return nil
}
