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

package keygen

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/security"
	"github.com/emitter-io/emitter/internal/security/hash"
	"github.com/emitter-io/emitter/internal/security/license"
)

// Provider represents a key generation provider.
type Provider struct {
	Cipher license.Cipher    // Cipher to use for the key generation
	Loader contract.Provider // Contract loader to use to retrieve contracts
}

// NewProvider creates a new key generation provider.
func NewProvider(cipher license.Cipher, loader contract.Provider) *Provider {
	return &Provider{
		Cipher: cipher,
		Loader: loader,
	}
}

// DecryptKey decrypts a key and returns it
func (p *Provider) DecryptKey(key string) (security.Key, error) {
	return p.Cipher.DecryptKey([]byte(key))
}

// EncryptKey encrypts the security key
func (p *Provider) EncryptKey(key security.Key) (string, error) {
	return p.Cipher.EncryptKey([]byte(key))
}

// CreateKey generates a key with the specified access and expiration time.
func (p *Provider) CreateKey(rawMasterKey, channel string, access uint8, expires time.Time) (string, *errors.Error) {
	masterKey, err := p.DecryptKey(rawMasterKey)
	if err != nil || !masterKey.IsMaster() || masterKey.IsExpired() {
		return "", errors.ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := p.Loader.Get(masterKey.Contract())
	if !contractFound {
		return "", errors.ErrNotFound
	}

	// Validate the contract
	if !contract.Validate(masterKey) {
		return "", errors.ErrUnauthorized
	}

	// Generate random salt
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt16))
	if err != nil {
		return "", errors.ErrServerError
	}

	// Create a key request
	key := security.Key(make([]byte, 24))
	key.SetSalt(uint16(n.Uint64()))
	key.SetMaster(masterKey.Master())
	key.SetContract(masterKey.Contract())
	key.SetSignature(masterKey.Signature())
	key.SetPermissions(access)
	key.SetExpires(expires)

	// Make sure we don't accidentally generate master keys
	key.SetPermission(security.AllowMaster, false)

	// Set the target and return an convert the error if it occurs
	if err := key.SetTarget(channel); err != nil {
		switch err {
		case security.ErrTargetInvalid:
			return "", errors.ErrTargetInvalid
		case security.ErrTargetTooLong:
			return "", errors.ErrTargetTooLong
		default:
			return "", errors.ErrServerError
		}
	}

	// Encrypt the final key
	out, err := p.Cipher.EncryptKey(key)
	if err != nil {
		return "", errors.ErrServerError
	}

	return out, nil
}

// ExtendKey creates a private channel and an appropriate key.
func (p *Provider) ExtendKey(channelKey, channelName, connectionID string, access uint8, expires time.Time) (*security.Channel, *errors.Error) {
	var suffix string
	if strings.HasSuffix(channelName, "#/") {
		suffix = "#/"
		channelName = strings.TrimSuffix(channelName, "#/")
	}

	channel := security.MakeChannel(channelKey, channelName)
	if channel.ChannelType != security.ChannelStatic {
		return nil, errors.ErrBadRequest
	}

	// Make sure we can actually extend it
	_, key, allowed := p.authorize(channel, security.AllowExtend)
	if !allowed {
		return nil, errors.ErrUnauthorized
	}

	// Revoke the extend permission to avoid this to be subsequently extended
	key.SetPermission(security.AllowExtend, false)

	// Apply the access and expiration
	key.SetPermissions(key.Permissions() & access)
	key.SetExpires(expires)

	// Create a new key for the private link
	target := fmt.Sprintf("%s%s/%s", channel.Channel, connectionID, suffix)
	if err := key.SetTarget(target); err != nil {
		return nil, errors.New(err.Error())
	}

	// Encrypt the key for storing
	encryptedKey, err := p.Cipher.EncryptKey(key)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	// Create the private channel
	channel.Channel = []byte(target)
	channel.Query = append(channel.Query, hash.OfString(connectionID))
	channel.Key = []byte(encryptedKey)
	return channel, nil
}

// Authorize attempts to authorize a channel with its key
func (p *Provider) authorize(channel *security.Channel, permission uint8) (contract.Contract, security.Key, bool) {

	// Attempt to parse the key
	key, err := p.Cipher.DecryptKey(channel.Key)
	if err != nil || key.IsExpired() {
		return nil, nil, false
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := p.Loader.Get(key.Contract())
	if !contractFound || !contract.Validate(key) || !key.HasPermission(permission) || !key.ValidateChannel(channel) {
		return nil, nil, false
	}

	// Return the contract and the key
	return contract, key, true
}
