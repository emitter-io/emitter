/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
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

package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math"
	"math/big"
	"strconv"
	"time"
)

const (
	xteaRounds = 32
	xteaDelta  = uint32(0x9E3779B9)
	xteaSum    = uint32(0xC6EF3720) // should be delta * rounds
)

// The map used for base64 decode.
var decodeMap [256]byte

type corruptInputError int64

func (e corruptInputError) Error() string {
	return "illegal base64 data at input byte " + strconv.FormatInt(int64(e), 10)
}

// Init prepares the lookup table.
func init() {
	encoder := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	for i := 0; i < len(decodeMap); i++ {
		decodeMap[i] = 0xFF
	}
	for i := 0; i < len(encoder); i++ {
		decodeMap[encoder[i]] = byte(i)
	}
}

// Cipher represents a security cipher which can encrypt/decrypt security keys.
type Cipher struct {
	key [4]uint32 // The cryptographic key used by the encryption algorithm.
}

// NewCipher creates a new cipher.
func NewCipher(value string) (*Cipher, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}

	if len(value) != 22 || len(data) != 16 {
		return nil, errors.New("Key provided is invalid")
	}

	cipher := new(Cipher)
	for i := 0; i < 4; i++ {
		cipher.key[i] = uint32((uint32(data[(4*i)+0]) << 24) |
			(uint32(data[(4*i)+1]) << 16) |
			(uint32(data[(4*i)+2]) << 8) |
			uint32(data[(4*i)+3]))
	}

	return cipher, nil
}

// DecryptKey decrypts the security key from a base64 encoded string.
func (c *Cipher) DecryptKey(buffer []byte) (Key, error) {
	if len(buffer) != 32 {
		return nil, errors.New("Key provided is invalid")
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
	return Key(buffer), nil
}

// EncryptKey encrypts the key and return a base-64 encoded string.
func (c *Cipher) EncryptKey(k Key) (string, error) {
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

// GenerateKey generates a new key.
func (c *Cipher) GenerateKey(masterKey Key, channel string, permissions uint8, expires time.Time, maxRandSalt int16) (string, error) {
	if maxRandSalt <= 0 {
		maxRandSalt = math.MaxInt16
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxRandSalt)))
	if err != nil {
		return "", err
	}

	key := Key(make([]byte, 24))
	key.SetSalt(uint16(n.Uint64()))
	key.SetMaster(masterKey.Master())
	key.SetContract(masterKey.Contract())
	key.SetSignature(masterKey.Signature())
	key.SetPermissions(permissions)
	key.SetExpires(expires)
	if err := key.SetTarget(channel); err != nil {
		return "", err
	}

	return c.EncryptKey(key)
}

// encrypt encrypts the data. This is done in-place and it's actually
// going to modify the underlying buffer.
func (c *Cipher) encrypt(data []byte) error {
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
func (c *Cipher) decrypt(data []byte) {
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

// decodeKey decodes the key from base64 string, url-encoded with no
// padding. This is 2x faster than the built-in function as we trimmed
// it significantly.
func decodeKey(dst, src []byte) (n int, err error) {
	var idx int
	for idx < len(src) {
		var dbuf [4]byte
		dinc, dlen := 3, 4

		for j := range dbuf {
			if len(src) == idx {
				if j < 2 {
					return n, corruptInputError(idx - j)
				}
				dinc, dlen = j-1, j
				break
			}

			in := src[idx]
			idx++

			dbuf[j] = decodeMap[in]
			if dbuf[j] == 0xFF {
				return n, corruptInputError(idx - 1)
			}
		}

		// Convert 4x 6bit source bytes into 3 bytes
		val := uint(dbuf[0])<<18 | uint(dbuf[1])<<12 | uint(dbuf[2])<<6 | uint(dbuf[3])
		dbuf[2], dbuf[1], dbuf[0] = byte(val>>0), byte(val>>8), byte(val>>16)
		switch dlen {
		case 4:
			dst[2] = dbuf[2]
			dbuf[2] = 0
			fallthrough
		case 3:
			dst[1] = dbuf[1]
			dbuf[1] = 0
			fallthrough
		case 2:
			dst[0] = dbuf[0]
		}

		dst = dst[dinc:]
		n += dlen - 1
	}

	return n, err
}
