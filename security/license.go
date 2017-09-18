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
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"time"
)

// Gets the beginning of time for the timestamp, which is 2010/1/1 00:00:00
const timeOffset = int64(1262304000)

// The beginning of time...
var timeZero = time.Unix(0, 0)

// Various license types
const (
	LicenseTypeUnknown = iota
	LicenseTypeCloud
	LicenseTypeOnPremise
)

// License represents a security license for the service.
type License struct {
	EncryptionKey string    // Gets or sets the encryption key.
	Contract      uint32    // Gets or sets the contract id.
	Signature     uint32    // Gets or sets the signature of the contract.
	Expires       time.Time // Gets or sets the expiration date for the license.
	Type          uint32    // Gets or sets the license type.
}

// NewLicense generates a new crypto-random license.
func NewLicense() *License {
	raw := make([]byte, 24)
	rand.Read(raw)

	return &License{
		EncryptionKey: base64.RawURLEncoding.EncodeToString(raw[0:16]),
		Contract:      uint32(binary.BigEndian.Uint32(raw[16:20])),
		Signature:     uint32(binary.BigEndian.Uint32(raw[20:24])),
		Expires:       time.Unix(0, 0),
		Type:          LicenseTypeOnPremise,
	}
}

// NewLicenseAndMaster generates a new license and master key.
func NewLicenseAndMaster() (string, string) {
	license := NewLicense()
	secret, err := license.NewMasterKey(1)
	if err != nil {
		panic(err)
	}

	cipher, err := license.Cipher()
	if err != nil {
		panic(err)
	}

	master, err := cipher.EncryptKey(secret)
	if err != nil {
		panic(err)
	}

	return license.String(), master
}

// ParseLicense decrypts the license and verifies it.
func ParseLicense(data string) (*License, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("No license was found, please provide a valid license key through the configuration file, an EMITTER_LICENSE environment variable or a valid vault key 'secrets/emitter/license'")
	}

	// Decode from base64 first
	raw, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// Get the expiration time
	expiry := int64(binary.BigEndian.Uint32(raw[24:28]))
	if expiry > 0 {
		expiry = timeOffset + expiry
	}

	// Parse the license
	license := License{
		EncryptionKey: base64.RawURLEncoding.EncodeToString(raw[0:16]),
		Contract:      uint32(binary.BigEndian.Uint32(raw[16:20])),
		Signature:     uint32(binary.BigEndian.Uint32(raw[20:24])),
		Expires:       time.Unix(expiry, 0),
		Type:          binary.BigEndian.Uint32(raw[28:32]),
	}

	return &license, nil
}

// NewMasterKey generates a new master key.
func (l *License) NewMasterKey(id uint16) (Key, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt16))
	if err != nil {
		return nil, err
	}

	key := Key(make([]byte, 24))
	key.SetSalt(uint16(n.Uint64()))
	key.SetMaster(id)
	key.SetContract(l.Contract)
	key.SetSignature(l.Signature)
	key.SetPermissions(AllowMaster)
	return key, nil
}

// Cipher creates a new cipher for the licence
func (l *License) Cipher() (*Cipher, error) {
	return NewCipher(l.EncryptionKey)
}

// String converts the license to string.
func (l *License) String() string {
	output := make([]byte, 32)
	key, err := base64.RawURLEncoding.DecodeString(l.EncryptionKey)
	if err != nil {
		return ""
	}

	expiry := l.Expires.Unix()
	if expiry > 0 {
		expiry = l.Expires.Unix() - timeOffset
	}

	copy(output, key)
	binary.BigEndian.PutUint32(output[16:20], uint32(l.Contract))
	binary.BigEndian.PutUint32(output[20:24], uint32(l.Signature))
	binary.BigEndian.PutUint32(output[24:28], uint32(expiry))
	binary.BigEndian.PutUint32(output[28:32], uint32(l.Type))
	return base64.RawURLEncoding.EncodeToString(output)
}
