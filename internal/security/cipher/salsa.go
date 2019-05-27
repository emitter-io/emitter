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

// Salsa represents a security cipher which can encrypt/decrypt security keys.
type Salsa struct {
	key   [32]byte
	nonce [24]byte
}

// NewSalsa creates a new salsa cipher.
func NewSalsa(key, nonce []byte) (*Salsa, error) {
	if len(key) != 32 || len(nonce) != 24 {
		return nil, errors.New("salsa: invalid cryptographic key")
	}

	cipher := new(Salsa)
	copy(cipher.key[:], key)
	copy(cipher.nonce[:], nonce)
	return cipher, nil
}

// setup produces a sub-key and Salsa20 counter given a nonce and key.
func (c *Salsa) setup(subKey *[32]byte, counter *[16]byte, nonce *[24]byte) {
	// We use XSalsa20 for encryption so first we need to generate a
	// key and nonce with HSalsa20.
	var hNonce [16]byte
	copy(hNonce[:], nonce[:])
	salsa.HSalsa20(subKey, &hNonce, &c.key, &salsa.Sigma)

	// The final 8 bytes of the original nonce form the new nonce.
	copy(counter[:], nonce[16:])
}

// EncryptKey encrypts the key and return a base-64 encoded string.
func (c *Salsa) EncryptKey(k security.Key) (string, error) {
	buffer := make([]byte, 24)
	copy(buffer[:], k)

	err := c.box(buffer)
	return base64.RawURLEncoding.EncodeToString(buffer), err
}

// DecryptKey decrypts the security key from a base64 encoded string.
func (c *Salsa) DecryptKey(buffer []byte) (security.Key, error) {
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

	// Warning: we do a XTEA decryption in same underlying buffer, to save up
	// on memory allocations. Keep in mind that the previous data will be lost.
	c.box(buffer)

	// Return the key on the decrypted buffer.
	return security.Key(buffer), nil
}

// box encrypts or decrypts the data. This is done in-place and it's actually
// going to modify the underlying buffer.
func (c *Salsa) box(data []byte) error {

	var subKey [32]byte
	var counter [16]byte
	c.setup(&subKey, &counter, &c.nonce)

	salsa.XORKeyStream(data, data, &counter, &subKey)
	return nil
}
