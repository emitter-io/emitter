// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package address

import (
	"math/rand"
	"net"
	"strings"
)

var hardware = initHardware()

// GetHardware retrieves the hardware address fingerprint which is based
// on the address of the first interface encountered.
func GetHardware() Fingerprint {
	return Fingerprint(hardware)
}

// Initializes the fingerprint
func initHardware() uint64 {
	var hardwareAddr [6]byte
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if len(iface.HardwareAddr) >= 6 {
				copy(hardwareAddr[:], iface.HardwareAddr)
				return encode(hardwareAddr[:])
			}
		}
	}

	safeRandom(hardwareAddr[:])
	hardwareAddr[0] |= 0x01
	return encode(hardwareAddr[:])
}

func encode(mac net.HardwareAddr) (r uint64) {
	for _, b := range mac {
		r <<= 8
		r |= uint64(b)
	}
	return
}

func safeRandom(dest []byte) {
	if _, err := rand.Read(dest); err != nil {
		panic(err)
	}
}

// Fingerprint represents hardware fingerprint
type Fingerprint uint64

// String encodes PeerName as a string.
func (f Fingerprint) String() string {
	return intmac(uint64(f)).String()
}

// Hex returns the string in hex format.
func (f Fingerprint) Hex() string {
	return strings.Replace(f.String(), ":", "", -1)
}

// Converts int to hardware address
func intmac(key uint64) (r net.HardwareAddr) {
	r = make([]byte, 6)
	for i := 5; i >= 0; i-- {
		r[i] = byte(key)
		key >>= 8
	}
	return
}
