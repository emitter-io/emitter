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
	"encoding/json"
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
	"github.com/emitter-io/emitter/internal/service"
)

// Service represents a key generation service.
type Service struct {
	cipher license.Cipher     // Cipher to use for the key generation
	loader contract.Provider  // Contract loader to use to retrieve contracts
	auth   service.Authorizer // The authorizer to use.
}

// New creates a new key generation provider.
func New(cipher license.Cipher, loader contract.Provider, auth service.Authorizer) *Service {
	return &Service{
		cipher: cipher,
		loader: loader,
		auth:   auth,
	}
}

// OnRequest processes a keygen request.
func (s *Service) OnRequest(c service.Conn, payload []byte) (service.Response, bool) {
	var message Request
	if err := json.Unmarshal(payload, &message); err != nil {
		return errors.ErrBadRequest, false
	}

	// Decrypt the parent key and make sure it's not expired
	parentKey, err := s.DecryptKey(message.Key)
	if err != nil || parentKey.IsExpired() {
		return errors.ErrUnauthorized, false
	}

	// If the key provided is a master key, create a new key
	if parentKey.IsMaster() {
		key, err := s.CreateKey(message.Key, message.Channel, message.access(), message.expires())
		if err != nil {
			return err, false
		}

		// Success, return the response
		return &Response{
			Status:  200,
			Key:     key,
			Channel: message.Channel,
		}, true
	}

	// If the key provided can be extended, attempt to extend the key
	if parentKey.HasPermission(security.AllowExtend) {
		channel, err := s.ExtendKey(message.Key, message.Channel, c.ID(), message.access(), message.expires())
		if err != nil {
			return err, false
		}

		// Success, return the response
		return &Response{
			Status:  200,
			Key:     string(channel.Key),     // Encrypted channel key
			Channel: string(channel.Channel), // Channel name
		}, true
	}

	// Not authorised
	return errors.ErrUnauthorized, false
}

// DecryptKey decrypts a key and returns it
func (s *Service) DecryptKey(key string) (security.Key, error) {
	return s.cipher.DecryptKey([]byte(key))
}

// EncryptKey encrypts the security key
func (s *Service) EncryptKey(key security.Key) (string, error) {
	return s.cipher.EncryptKey([]byte(key))
}

// CreateKey generates a key with the specified access and expiration time.
func (s *Service) CreateKey(rawMasterKey, channel string, access uint8, expires time.Time) (string, *errors.Error) {
	masterKey, err := s.DecryptKey(rawMasterKey)
	if err != nil || !masterKey.IsMaster() || masterKey.IsExpired() {
		return "", errors.ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := s.loader.Get(masterKey.Contract())
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
	out, err := s.cipher.EncryptKey(key)
	if err != nil {
		return "", errors.ErrServerError
	}

	return out, nil
}

// ExtendKey creates a private channel and an appropriate key.
func (s *Service) ExtendKey(channelKey, channelName, connectionID string, access uint8, expires time.Time) (*security.Channel, *errors.Error) {
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
	_, key, allowed := s.auth.Authorize(channel, security.AllowExtend)
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
	encryptedKey, err := s.cipher.EncryptKey(key)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	// Create the private channel
	channel.Channel = []byte(target)
	channel.Query = append(channel.Query, hash.OfString(connectionID))
	channel.Key = []byte(encryptedKey)
	return channel, nil
}
