package broker

import (
	"github.com/emitter-io/emitter/security"
	"math"
	"math/rand"
)

const (
	RequestKeygen   = 548658350
	RequestPresence = 3869262148
)

const (
	QueryType          = 1
	QueryActualKey     = 2 // As opposed to 'emitter/'
	QueryTargetChannel = 3
)

// TryProcessAPIRequest attempts to generate the key and returns the result.
func TryProcessAPIRequest(channel *security.Channel) bool {
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
		return true
	case RequestPresence:
		return true
	default:
		return true
	}

}

func ProcessKeyGen(channel *security.Channel) bool {
	// Attempt to parse the key, this should be a master key

	// Attempt to fetch the contract using the key. Underneath, it's cached.

	// Validate the contract

	// Generate the key
	key := security.Key(make([]byte, 24))
	key.SetSalt(uint16(rand.Intn(math.MaxInt16)))
	key.SetMaster(2)
	key.SetContract(123)
	key.SetSignature(777)
	//key.SetPermissions(AllowReadWrite)
	key.SetTarget(56789)
	//key.SetExpires(time.Unix(1497683272, 0).UTC())
	return false
}
