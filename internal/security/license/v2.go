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

	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/security/cipher"
	"github.com/golang/snappy"
	"github.com/kelindar/binary"
)

// V2 represents a v2 license.
type V2 struct {
	EncryptionKey  []byte // Gets or sets the encryption key.
	EncryptionSalt []byte // Gets or sets the encryption key.
	User           uint32 // Gets or sets the contract id.
	Sign           uint32 // Gets or sets the signature of the contract.
	Index          uint32 // Gets or sets the current master.
}

// NewV2 generates a new v2 license.
func NewV2() *V2 {
	return &V2{
		EncryptionKey:  randN(32),
		EncryptionSalt: randN(24),
		User:           uint32(be.Uint32(randN(4))),
		Sign:           uint32(be.Uint32(randN(4))),
		Index:          1,
	}
}

// parseV2 decodes the license and verifies it.
func parseV2(data string) (*V2, error) {

	// Decode from base64 first
	raw, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// Uncompress the bytes
	raw, err = snappy.Decode(nil, raw)
	if err != nil {
		return nil, err
	}

	// Unmarshal the license
	var license V2
	err = binary.Unmarshal(raw, &license)
	return &license, err
}

// Cipher creates a new cipher for the licence
func (l *V2) Cipher() (Cipher, error) {
	return cipher.NewSalsa(l.EncryptionKey, l.EncryptionSalt)
}

// String converts the license to string.
func (l *V2) String() string {
	encoded, _ := binary.Marshal(l)
	encoded = snappy.Encode(nil, encoded)
	return base64.RawURLEncoding.EncodeToString(encoded) + ":2"
}

// Contract retuns the contract ID of the license.
func (l *V2) Contract() uint32 {
	return l.User
}

// Signature returns the signature of the license.
func (l *V2) Signature() uint32 {
	return l.Sign
}

// Master returns the secret key index.
func (l *V2) Master() uint32 {
	return l.Index
}

// NewMasterKey generates a new master key.
func (l *V2) NewMasterKey(id uint16) (key security.Key, err error) {
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
