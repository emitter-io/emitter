// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package address

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Resolve resolves the addresses behind a hostname or a URL. This supports
// only TCP (as pretty much everything in the package) but can handle a URL
// passed instead of the name (e.g. https://example.com).
func Resolve(name string, defaultPort int) ([]net.TCPAddr, error) {
	if name == "" {
		return nil, errors.New("unable to resolve an empty name or url")
	}

	// First we need to check whether the address is a valid IPv4 or IPv6 address
	if ip := net.ParseIP(name); ip != nil {
		return []net.TCPAddr{{IP: ip, Port: defaultPort}}, nil
	}

	// If this is not a url, add a scheme to make it a URL.
	if !strings.Contains(name, "://") {
		name = "tcp://" + name
	}

	// Check if this is a valid url, then adjust port and hostname
	port := defaultPort
	url, err := url.Parse(name)
	if err != nil {
		return nil, err
	}

	// The hostname might not be there (if it's an address)
	if url.Hostname() != "" {
		name = url.Hostname()
	}

	// If we have a port on the url, use that as well
	if url.Port() != "" {
		urlPort, err := strconv.Atoi(url.Port())
		if err != nil {
			return nil, err
		}
		port = urlPort
	}

	// If we can parse the address, we can just return ip/port combination
	if ip := net.ParseIP(name); ip != nil {
		return []net.TCPAddr{{IP: ip, Port: port}}, nil
	}

	// Get the addresses by performing a DNS lookup, this should not fail
	addr, err := net.LookupHost(name)
	if err != nil {
		return nil, err
	}

	// Add port to each address
	var addrs []net.TCPAddr
	for _, a := range addr {
		if resolved, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", a, port)); err == nil {
			addrs = append(addrs, *resolved)
		}
	}
	return addrs, nil
}
