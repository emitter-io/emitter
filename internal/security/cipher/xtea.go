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
)

const (
	xteaRounds = 32
	xteaDelta  = uint32(0x9E3779B9)
	xteaSum    = uint32(0xC6EF3720) // should be delta * rounds
)

// Xtea represents a security cipher which can encrypt/decrypt security keys.
type Xtea struct {
	key [4]uint32 // The cryptographic key used by the encryption algorithm.
}

// NewXtea creates a new cipher.
func NewXtea(value string) (*Xtea, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}

	if len(value) != 22 || len(data) != 16 {
		return nil, errors.New("xtea: invalid cryptographic key")
	}

	cipher := new(Xtea)
	for i := 0; i < 4; i++ {
		cipher.key[i] = uint32((uint32(data[(4*i)+0]) << 24) |
			(uint32(data[(4*i)+1]) << 16) |
			(uint32(data[(4*i)+2]) << 8) |
			uint32(data[(4*i)+3]))
	}

	return cipher, nil
}

// DecryptKey decrypts the security key from a base64 encoded string.
func (c *Xtea) DecryptKey(buffer []byte) (security.Key, error) {
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
	buffer = buffer[0:n]

	// Warning: we do a XTEA decryption in same underlying buffer, to save up
	// on memory allocations. Keep in mind that the previous data will be lost.
	c.decrypt(buffer)

	// Then XOR the entire array with the salt. We alter the underlying buffer
	// for the 3rd time.
	for i := 2; i < 24; i += 2 {
		buffer[i] = byte(buffer[i] ^ buffer[0])
		buffer[i+1] = byte(buffer[i+1] ^ buffer[1])
	}

	// Return the key on the decrypted buffer.
	return security.Key(buffer), nil
}

// EncryptKey encrypts the key and return a base-64 encoded string.
func (c *Xtea) EncryptKey(k security.Key) (string, error) {
	buffer := make([]byte, 24)
	buffer[0] = k[0]
	buffer[1] = k[1]

	// First XOR the entire array with the salt
	for i := 2; i < 24; i += 2 {
		buffer[i] = byte(k[i] ^ buffer[0])
		buffer[i+1] = byte(k[i+1] ^ buffer[1])
	}

	// Then encrypt the key using the master key
	//fmt.Printf("%v", buffer)
	err := c.encrypt(buffer)
	return base64.RawURLEncoding.EncodeToString(buffer), err
}

// encrypt encrypts the data. This is done in-place and it's actually
// going to modify the underlying buffer.
func (c *Xtea) encrypt(data []byte) error {
	if len(data) != 24 {
		return errors.New("The security key should be 24-bytes long")
	}

	key := &c.key
	var r, sum, y, z uint32
	for i := 0; i < len(data); i += 8 {
		y = uint32(data[i])<<24 | uint32(data[i+1])<<16 | uint32(data[i+2])<<8 | uint32(data[i+3])
		z = uint32(data[i+4])<<24 | uint32(data[i+5])<<16 | uint32(data[i+6])<<8 | uint32(data[i+7])

		// Encipher the block
		sum = 0
		for r = 0; r < xteaRounds; r++ {
			y += (((z << 4) ^ (z >> 5)) + z) ^ (sum + key[sum&3])
			sum += xteaDelta
			z += (((y << 4) ^ (y >> 5)) + y) ^ (sum + key[(sum>>11)&3])
		}

		// Set to the current block
		data[i] = (byte)(y >> 24)
		data[i+1] = (byte)(y >> 16)
		data[i+2] = (byte)(y >> 8)
		data[i+3] = (byte)(y)
		data[i+4] = (byte)(z >> 24)
		data[i+5] = (byte)(z >> 16)
		data[i+6] = (byte)(z >> 8)
		data[i+7] = (byte)(z)
	}

	return nil
}

// decrypt decrypts the data. This is done in-place and it's actually
// going to modify the underlying buffer.
func (c *Xtea) decrypt(data []byte) {
	key := &c.key
	for i := 0; i < 24; i += 8 {
		y := uint32(data[i])<<24 | uint32(data[i+1])<<16 | uint32(data[i+2])<<8 | uint32(data[i+3])
		z := uint32(data[i+4])<<24 | uint32(data[i+5])<<16 | uint32(data[i+6])<<8 | uint32(data[i+7])

		// Decipher the block
		sum := xteaSum
		for r := 0; r < xteaRounds; r++ {
			z -= (((y << 4) ^ (y >> 5)) + y) ^ (sum + key[(sum>>11)&3])
			sum -= xteaDelta
			y -= (((z << 4) ^ (z >> 5)) + z) ^ (sum + key[sum&3])
		}

		// Set to the current block
		data[i] = (byte)(y >> 24)
		data[i+1] = (byte)(y >> 16)
		data[i+2] = (byte)(y >> 8)
		data[i+3] = (byte)(y)
		data[i+4] = (byte)(z >> 24)
		data[i+5] = (byte)(z >> 16)
		data[i+6] = (byte)(z >> 8)
		data[i+7] = (byte)(z)
	}
}
