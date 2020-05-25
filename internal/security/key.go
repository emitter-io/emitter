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

package security

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/emitter-io/emitter/internal/security/hash"
	"github.com/kelindar/binary"
)

// Gets the beginning of time for the timestamp, which is 2010/1/1 00:00:00
const timeOffset = int64(1262304000)

// The beginning of time...
var timeZero = time.Unix(0, 0)

// Access types for a security key.
const (
	AllowNone      = uint8(0)                     // Key has no privileges.
	AllowMaster    = uint8(1 << 0)                // Key should be allowed to generate other keys.
	AllowRead      = uint8(1 << 1)                // Key should be allowed to subscribe to the target channel.
	AllowWrite     = uint8(1 << 2)                // Key should be allowed to publish to the target channel.
	AllowStore     = uint8(1 << 3)                // Key should be allowed to write to the message history of the target channel.
	AllowLoad      = uint8(1 << 4)                // Key should be allowed to write to read the message history of the target channel.
	AllowPresence  = uint8(1 << 5)                // Key should be allowed to query the presence on the target channel.
	AllowExtend    = uint8(1 << 6)                // Key should be allowed to create sub-channels by extending an existing one.
	AllowExecute   = uint8(1 << 7)                // Key should be allowed to execute code. (RESERVED)
	AllowReadWrite = AllowRead | AllowWrite       // Key should be allowed to read and write to the target channel.
	AllowStoreLoad = AllowStore | AllowLoad       // Key should be allowed to read and write the message history.
	AllowAll       = math.MaxUint8 &^ AllowMaster // Key allows everything except master
)

// Key errors
var (
	ErrTargetInvalid = errors.New("channel should end with `/` for strict types or `/#/` for multi level wildcard")
	ErrTargetTooLong = errors.New("channel can not have more than 23 parts")
)

// Key represents a security key.
type Key []byte

// IsEmpty checks whether the key is empty or not.
func (k Key) IsEmpty() bool {
	return len(k) == 0
}

// Salt gets the random salt of the key
func (k Key) Salt() uint16 {
	return uint16(k[0])<<8 | uint16(k[1])
}

// SetSalt sets the random salt of the key.
func (k Key) SetSalt(value uint16) {
	k[0] = byte(value >> 8)
	k[1] = byte(value)
}

// Master gets the master key id.
func (k Key) Master() uint16 {
	return uint16(k[2])<<8 | uint16(k[3])
}

// SetMaster sets the master key id.
func (k Key) SetMaster(value uint16) {
	k[2] = byte(value >> 8)
	k[3] = byte(value)
}

// Contract gets the contract id.
func (k Key) Contract() uint32 {
	return uint32(k[4])<<24 | uint32(k[5])<<16 | uint32(k[6])<<8 | uint32(k[7])
}

// SetContract sets the contract id.
func (k Key) SetContract(value uint32) {
	k[4] = byte(value >> 24)
	k[5] = byte(value >> 16)
	k[6] = byte(value >> 8)
	k[7] = byte(value)
}

// Signature gets the signature of the contract.
func (k Key) Signature() uint32 {
	return uint32(k[8])<<24 | uint32(k[9])<<16 | uint32(k[10])<<8 | uint32(k[11])
}

// SetSignature sets the signature of the contract.
func (k Key) SetSignature(value uint32) {
	k[8] = byte(value >> 24)
	k[9] = byte(value >> 16)
	k[10] = byte(value >> 8)
	k[11] = byte(value)
}

// Permissions gets the permission flags.
func (k Key) Permissions() uint8 {
	return k[15]
}

// SetPermissions sets the permission flags.
func (k Key) SetPermissions(value uint8) {
	k[15] = value
}

// ValidateChannel validates the channel string.
func (k Key) ValidateChannel(ch *Channel) bool {
	topic := ch.Channel
	if len(topic) == 0 {
		return false
	}

	// Bytes 16-17-18-19 contains target hash
	target := uint32(k[16])<<24 | uint32(k[17])<<16 | uint32(k[18])<<8 | uint32(k[19])
	targetPath := uint32(k[12])<<16 | uint32(k[13])<<8 | uint32(k[14])

	// Retro-compatibility: if there's no depth specified we default to a single-level validation
	if targetPath == 0 {
		if target == 1325880984 { // Key target was "#/" (1325880984 == hash(""))
			return true
		}
		return target == ch.Target()
	}

	// Trim right `/`
	if topic[len(topic)-1] == '/' {
		topic = topic[:len(topic)-1]
	}

	// Split by `/`
	channel := binary.ToString(&topic)
	parts := strings.Split(channel, "/")
	wc := parts[len(parts)-1] == "#"
	if wc {
		parts = parts[0 : len(parts)-1]
	}

	maxDepth := 0
	for i := uint32(0); i < 23; i++ {
		if ((targetPath >> i) & 1) == 1 {
			maxDepth = 23 - int(i)
			break
		}
	}

	// If no depth defined, all the parts in key target were wildcards (+)
	// We need to compare the key hash with the whole channel we received.
	if maxDepth == 0 {
		maxDepth = len(parts)
	}

	// Get the first bit, whether the key is the exact match or not
	keyIsExactTarget := ((targetPath >> 23) & 1) == 1
	if len(parts) < maxDepth || (keyIsExactTarget && len(parts) != maxDepth) {
		return false
	}

	for idx, part := range parts {
		if ((targetPath >> (22 - uint32(idx))) & 1) == 1 {
			if part == "+" {
				return false
			}
		} else {
			parts[idx] = "+"
		}
	}

	newChannel := strings.Join(parts[0:maxDepth], "/")

	h := hash.OfString(newChannel)
	return h == target
}

// SetTarget sets the target channel for the key.
func (k Key) SetTarget(channel string) error {
	if !strings.HasSuffix(channel, "/") {
		return ErrTargetInvalid
	}

	// Get all of the parts for the target channel
	// History: https://github.com/emitter-io/emitter/issues/76
	parts := strings.Split(strings.TrimRight(channel, "/"), "/")
	wildcard := parts[len(parts)-1] == "#"

	// 1st bit is 0 for wildcard, 1 for strict type
	bitPath := uint32(1 << 23)
	if wildcard {
		parts = parts[0 : len(parts)-1]
		bitPath = 0
	}

	// Perform some validation
	if len(parts) > 23 {
		return ErrTargetTooLong
	}

	// Encode all of the parts
	for idx, part := range parts {
		if part != "+" && part != "#" {
			bitPath |= uint32(1 << (22 - uint16(idx)))
		}
	}

	// Create a new channel and get the hash for this channel
	newChannel := strings.Join(parts, "/")
	value := hash.OfString(newChannel)

	// Set the bit path
	k[12] = byte(bitPath >> 16)
	k[13] = byte(bitPath >> 8)
	k[14] = byte(bitPath)

	// Set the hash of the target
	k[16] = byte(value >> 24)
	k[17] = byte(value >> 16)
	k[18] = byte(value >> 8)
	k[19] = byte(value)
	return nil
}

// Expires gets the expiration date for the key.
func (k Key) Expires() time.Time {
	expire := int64(uint32(k[20])<<24 | uint32(k[21])<<16 | uint32(k[22])<<8 | uint32(k[23]))
	if expire > 0 {
		expire = timeOffset + expire
	}

	return time.Unix(expire, 0).UTC()
}

// SetExpires sets the expiration date for the key.
func (k Key) SetExpires(value time.Time) {
	expire := value.Unix()
	if expire > 0 {
		expire = expire - timeOffset
	}
	k[20] = byte(uint32(expire) >> 24)
	k[21] = byte(uint32(expire) >> 16)
	k[22] = byte(uint32(expire) >> 8)
	k[23] = byte(uint32(expire))
}

// IsExpired gets whether the key has expired or not.
func (k Key) IsExpired() bool {
	expiry := k.Expires()
	if expiry.Equal(timeZero) {
		return false
	}

	return expiry.Before(time.Now().UTC())
}

// IsMaster gets whether the key is a master key..
func (k Key) IsMaster() bool {
	return k.Permissions() == AllowMaster
}

// HasPermission check whether the key provides some permission.
func (k Key) HasPermission(flag uint8) bool {
	p := k.Permissions()
	return (p & flag) == flag
}

// SetPermission sets a permission to a value
func (k Key) SetPermission(flag uint8, value bool) {
	if value {
		k.SetPermissions(k.Permissions() | flag)
	} else {
		k.SetPermissions(k.Permissions() &^ flag)
	}
}
