// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package address

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
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

// GetPrivate returns the list of private network IPv4 addresses on
// all active interfaces.
func GetPrivate() ([]net.IPAddr, error) {
	addresses, err := activeInterfaceAddresses()
	if err != nil {
		return nil, fmt.Errorf("failed to get interface addresses: %v", err)
	}

	var addrs []net.IPAddr
	for _, rawAddr := range addresses {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}
		if ip.To4() == nil {
			continue
		}
		if !isPrivate(ip) {
			continue
		}
		addrs = append(addrs, net.IPAddr{IP: ip})
	}
	return addrs, nil
}

// GetPrivateOrDefault returns the first private IPv4 address. If no address
// is available, this defaults to the specified default address.
func GetPrivateOrDefault(addr net.IPAddr) net.IPAddr {
	if addrs, err := GetPrivate(); err == nil && len(addrs) > 0 {
		addr = addrs[0]
	}
	return addr
}

// GetPublic returns the list of all public IP addresses
// on all active interfaces.
func GetPublic() ([]net.IPAddr, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to get interface addresses: %v", err)
	}

	var addrs []net.IPAddr
	for _, rawAddr := range addresses {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}
		if isPrivate(ip) {
			continue
		}
		addrs = append(addrs, net.IPAddr{IP: ip})
	}
	return addrs, nil
}

// GetPublicOrDefault returns the first public IP address. If no address
// is available, this defaults to the specified default address.
func GetPublicOrDefault(addr net.IPAddr) net.IPAddr {
	if addrs, err := GetPublic(); err == nil && len(addrs) > 0 {
		addr = addrs[0]
	}
	return addr
}

// GetExternal returns an externally visible IP Address. This actually
// goes out to the internet service which echoes the address.
func GetExternal() (net.IPAddr, error) {
	ip, ok := getExternal(
		"http://ipv4.icanhazip.com",
		"http://myexternalip.com/raw",
		"http://www.trackip.net/ip",
		"http://automation.whatismyip.com/n09230945.asp",
		"http://api.ipify.org/",
	)

	// Fallback to localhost
	if !ok || ip == nil {
		return net.IPAddr{}, fmt.Errorf("failed to get external address")
	}

	return net.IPAddr{IP: ip}, nil
}

// GetExternalOrDefault returns an externally visible IP Address. If no address
// is available, this defaults to the specified default address.
func GetExternalOrDefault(defaultAddr net.IPAddr) net.IPAddr {
	if addr, err := GetExternal(); err == nil {
		return addr
	}
	return defaultAddr
}

// privateBlocks contains non-forwardable address blocks which are used
// for private networks. RFC 6890 provides an overview of special
// address blocks.
var privateBlocks = []*net.IPNet{
	parseCIDR("10.0.0.0/8"),     // RFC 1918 IPv4 private network address
	parseCIDR("100.64.0.0/10"),  // RFC 6598 IPv4 shared address space
	parseCIDR("127.0.0.0/8"),    // RFC 1122 IPv4 loopback address
	parseCIDR("169.254.0.0/16"), // RFC 3927 IPv4 link local address
	parseCIDR("172.16.0.0/12"),  // RFC 1918 IPv4 private network address
	parseCIDR("192.0.0.0/24"),   // RFC 6890 IPv4 IANA address
	parseCIDR("192.0.2.0/24"),   // RFC 5737 IPv4 documentation address
	parseCIDR("192.168.0.0/16"), // RFC 1918 IPv4 private network address
	parseCIDR("::1/128"),        // RFC 1884 IPv6 loopback address
	parseCIDR("fe80::/10"),      // RFC 4291 IPv6 link local addresses
	parseCIDR("fc00::/7"),       // RFC 4193 IPv6 unique local addresses
	parseCIDR("fec0::/10"),      // RFC 1884 IPv6 site-local addresses
	parseCIDR("2001:db8::/32"),  // RFC 3849 IPv6 documentation address
}

func parseCIDR(s string) *net.IPNet {
	_, block, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Sprintf("Bad CIDR %s: %s", s, err))
	}
	return block
}

func isPrivate(ip net.IP) bool {
	for _, priv := range privateBlocks {
		if priv.Contains(ip) {
			return true
		}
	}
	return false
}

// Returns addresses from interfaces that is up
func activeInterfaceAddresses() ([]net.Addr, error) {
	var upAddrs []net.Addr
	var loAddrs []net.Addr

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("Failed to get interfaces: %v", err)
	}

	for _, iface := range interfaces {
		// Require interface to be up
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("Failed to get interface addresses: %v", err)
		}

		if iface.Flags&net.FlagLoopback != 0 {
			loAddrs = append(loAddrs, addresses...)
			continue
		}

		upAddrs = append(upAddrs, addresses...)
	}

	if len(upAddrs) == 0 {
		return loAddrs, nil
	}

	return upAddrs, nil
}

// getExternal retrieves an external IP address
func getExternal(urls ...string) (net.IP, bool) {
	for _, url := range urls {
		cli := http.Client{Timeout: time.Duration(5 * time.Second)}
		res, err := cli.Get(url)
		if err == nil {

			// Read the response
			defer res.Body.Close()
			r, err := ioutil.ReadAll(res.Body)
			if err == nil {

				// Fix and parse
				addr := strings.Replace(string(r), "\n", "", -1)
				ip := net.ParseIP(addr)
				return ip, ip != nil
			}
		}
	}

	return nil, false
}
