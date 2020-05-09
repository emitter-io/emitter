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
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/emitter-io/emitter/internal/security"
)

var be = binary.BigEndian

// Cipher represents a cipher used by the license type.
type Cipher interface {
	DecryptKey(buffer []byte) (security.Key, error)
	EncryptKey(k security.Key) (string, error)
}

// License represents an abstract license.
type License interface {
	fmt.Stringer
	NewMasterKey(id uint16) (security.Key, error)
	Cipher() (Cipher, error)
	Contract() uint32
	Signature() uint32
	Master() uint32
}

// New generates a new license and master key. This uses the most up-to-date version
// of the license to generate a new one.
func New() (string, string) {
	license := NewV3()
	if secret, err := license.NewMasterKey(1); err != nil {
		panic(err)
	} else if cipher, err := license.Cipher(); err != nil {
		panic(err)
	} else if master, err := cipher.EncryptKey(secret); err != nil {
		panic(err)
	} else {
		return license.String(), master
	}
}

// Parse parses a valid license of any version.
func Parse(data string) (License, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("No license was found, please provide a valid license key through the configuration file, an EMITTER_LICENSE environment variable or a valid vault key 'secrets/emitter/license'")
	}

	switch {
	case strings.HasSuffix(data, ":1"):
		return parseV1(data[:len(data)-2])
	case strings.HasSuffix(data, ":2"):
		return parseV2(data[:len(data)-2])
	case strings.HasSuffix(data, ":3"):
		return parseV3(data[:len(data)-2])
	default:
		return parseV1(data)
	}
}

// RandN generates a crypto-random N bytes.
func randN(n int) []byte {
	raw := make([]byte, n)
	rand.Read(raw)
	return raw
}
