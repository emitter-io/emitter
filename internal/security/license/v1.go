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

package license

import (
	"crypto/rand"
	"encoding/base64"
	"math"
	"math/big"
	"time"

	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/security/cipher"
)

// Gets the beginning of time for the timestamp, which is 2010/1/1 00:00:00
const timeOffset = int64(1262304000)

// Various license types
const (
	LicenseTypeUnknown = iota
	LicenseTypeCloud
	LicenseTypeOnPremise
)

// V1 represents a legacy v1 license.
type V1 struct {
	EncryptionKey string    // Gets or sets the encryption key.
	User          uint32    // Gets or sets the contract id.
	Sign          uint32    // Gets or sets the signature of the contract.
	Expires       time.Time // Gets or sets the expiration date for the license.
	Type          uint32    // Gets or sets the license type.
}

// NewV1 generates a new legacy v1 license.
func NewV1() *V1 {
	return &V1{
		EncryptionKey: base64.RawURLEncoding.EncodeToString(randN(16)),
		User:          uint32(be.Uint32(randN(4))),
		Sign:          uint32(be.Uint32(randN(4))),
		Expires:       time.Unix(0, 0),
		Type:          LicenseTypeOnPremise,
	}
}

// parseV1 decrypts the license and verifies it.
func parseV1(data string) (*V1, error) {

	// Decode from base64 first
	raw, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// Get the expiration time
	expiry := int64(be.Uint32(raw[24:28]))
	if expiry > 0 {
		expiry = timeOffset + expiry
	}

	// Parse the license
	license := V1{
		EncryptionKey: base64.RawURLEncoding.EncodeToString(raw[0:16]),
		User:          uint32(be.Uint32(raw[16:20])),
		Sign:          uint32(be.Uint32(raw[20:24])),
		Expires:       time.Unix(expiry, 0),
		Type:          be.Uint32(raw[28:32]),
	}

	return &license, nil
}

// NewMasterKey generates a new master key.
func (l *V1) NewMasterKey(id uint16) (key security.Key, err error) {
	var n *big.Int
	if n, err = rand.Int(rand.Reader, big.NewInt(math.MaxInt16)); err == nil {
		key = security.Key(make([]byte, 24))
		key.SetSalt(uint16(n.Uint64()))
		key.SetMaster(id)
		key.SetContract(l.User)
		key.SetSignature(l.Sign)
		key.SetPermissions(security.AllowMaster)
	}
	return
}

// Cipher creates a new cipher for the licence
func (l *V1) Cipher() (Cipher, error) {
	return cipher.NewXtea(l.EncryptionKey)
}

// String converts the license to string.
func (l *V1) String() string {
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
	be.PutUint32(output[16:20], uint32(l.User))
	be.PutUint32(output[20:24], uint32(l.Sign))
	be.PutUint32(output[24:28], uint32(expiry))
	be.PutUint32(output[28:32], uint32(l.Type))
	return base64.RawURLEncoding.EncodeToString(output) + ":1"
}

// Contract retuns the contract ID of the license.
func (l *V1) Contract() uint32 {
	return l.User
}

// Signature returns the signature of the license.
func (l *V1) Signature() uint32 {
	return l.Sign
}

// Master returns the secret key index.
func (l *V1) Master() uint32 {
	return 1
}
