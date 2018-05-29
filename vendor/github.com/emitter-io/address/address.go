// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package address

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Loopback is the default loopback address (127.0.0.1)
var Loopback = net.IPAddr{IP: net.ParseIP("127.0.0.1")}

// Parse parses a TCP address + port combination.
func Parse(addr string, defaultPort int) (*net.TCPAddr, error) {
	if addr == "" {
		return nil, errors.New("unable to parse an empty address")
	}

	// Default to 0.0.0.0 interface
	if addr[0] == ':' {
		addr = "0.0.0.0" + addr
	}

	// Convenience: set private address
	if strings.Contains(addr, "private") {
		private := GetPrivateOrDefault(Loopback)
		addr = strings.Replace(addr, "private", private.String(), 1)
	}

	// Convenience: set public address
	if strings.Contains(addr, "external") {
		external := GetExternalOrDefault(Loopback)
		addr = strings.Replace(addr, "external", external.String(), 1)
	}

	// Convenience: set public address
	if strings.Contains(addr, "public") {
		public := GetPublicOrDefault(Loopback)
		addr = strings.Replace(addr, "public", public.String(), 1)
	}

	// If we have only an IP address, use the default port
	if ip := net.ParseIP(addr); ip != nil {
		addr = fmt.Sprintf("%s:%d", ip, defaultPort)
	}

	// Resolve the address
	return net.ResolveTCPAddr("tcp", addr)
}
