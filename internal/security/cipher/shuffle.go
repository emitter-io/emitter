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

package cipher

import (
	"encoding/base64"
	"errors"

	"github.com/emitter-io/emitter/internal/security"
	"golang.org/x/crypto/salsa20/salsa"
)

// Shuffle represents a security cipher which can encrypt/decrypt security keys.
type Shuffle struct {
	key   [32]byte
	nonce [16]byte
}

// NewShuffle creates a new shuffled salsa cipher.
func NewShuffle(key, nonce []byte) (*Shuffle, error) {
	if len(key) != 32 || len(nonce) != 16 {
		return nil, errors.New("shuffled: invalid cryptographic key")
	}

	cipher := new(Shuffle)
	copy(cipher.key[:], key)
	copy(cipher.nonce[:], nonce)
	return cipher, nil
}

// EncryptKey encrypts the key and return a base-64 encoded string.
func (c *Shuffle) EncryptKey(k security.Key) (string, error) {
	buffer := make([]byte, 24)
	copy(buffer[:], k)

	err := c.crypt(buffer)
	return base64.RawURLEncoding.EncodeToString(buffer), err
}

// DecryptKey decrypts the security key from a base64 encoded string.
func (c *Shuffle) DecryptKey(buffer []byte) (security.Key, error) {
	if len(buffer) != 32 {
		return nil, errors.New("cipher: the key provided is not valid")
	}

	// Warning: we do a base64 decode in the same underlying buffer, to save up
	// on memory allocations. Keep in mind that the previous data will be lost.
	n, err := decodeKey(buffer, buffer)
	if err != nil {
		return nil, err
	}

	// We now need to resize the slice, since we changed it.
	buffer = buffer[:n]
	c.crypt(buffer)

	// Return the key on the decrypted buffer.
	return security.Key(buffer), nil
}

// crypt encrypts or decrypts the data and shuffles (recommended).
func (c *Shuffle) crypt(data []byte) error {
	buffer := data[2:]
	salt := data[0:2]

	// Apply the salt to nonce
	var nonce [16]byte
	for i := 0; i < 16; i += 2 {
		nonce[i] = salt[0] ^ c.nonce[i]
		nonce[i+1] = salt[1] ^ c.nonce[i+1]
	}

	var subKey [32]byte
	salsa.HSalsa20(&subKey, &nonce, &c.key, &salsa.Sigma)
	salsa.XORKeyStream(buffer, buffer, &nonce, &subKey)
	return nil
}
