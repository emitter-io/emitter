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
	"time"
)

// Access types for a security key.
const (
	AllowNone      = uint32(0)              // Key has no privileges.
	AllowMaster    = uint32(1 << 0)         // Key should be allowed to generate other keys.
	AllowRead      = uint32(1 << 1)         // Key should be allowed to subscribe to the target channel.
	AllowWrite     = uint32(1 << 2)         // Key should be allowed to publish to the target channel.
	AllowStore     = uint32(1 << 3)         // Key should be allowed to write to the message history of the target channel.
	AllowLoad      = uint32(1 << 4)         // Key should be allowed to write to read the message history of the target channel.
	AllowPresence  = uint32(1 << 5)         // Key should be allowed to query the presence on the target channel.
	AllowReadWrite = AllowRead | AllowWrite // Key should be allowed to read and write to the target channel.
	AllowStoreLoad = AllowStore | AllowLoad // Key should be allowed to read and write the message history.
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
func (k Key) Permissions() uint32 {
	return uint32(k[12])<<24 | uint32(k[13])<<16 | uint32(k[14])<<8 | uint32(k[15])
}

// SetPermissions sets the permission flags.
func (k Key) SetPermissions(value uint32) {
	k[12] = byte(value >> 24)
	k[13] = byte(value >> 16)
	k[14] = byte(value >> 8)
	k[15] = byte(value)
}

// Target gets the target for the key.
func (k Key) Target() uint32 {
	return uint32(k[16])<<24 | uint32(k[17])<<16 | uint32(k[18])<<8 | uint32(k[19])
}

// SetTarget sets the target for the key.
func (k Key) SetTarget(value uint32) {
	k[16] = byte(value >> 24)
	k[17] = byte(value >> 16)
	k[18] = byte(value >> 8)
	k[19] = byte(value)
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
func (k Key) HasPermission(flag uint32) bool {
	p := k.Permissions()
	return (p & flag) == flag
}
