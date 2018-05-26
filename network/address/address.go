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

package address

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/emitter-io/emitter/provider/logging"
	"github.com/hashicorp/go-sockaddr"
)

var hardware uint64
var external net.IP

func init() {
	external = initExternal()
	hardware = initHardware()
}

// External gets the external IP address.
func External() net.IP {
	return external
}

// Hardware gets the hardware address.
func Hardware() Fingerprint {
	return Fingerprint(hardware)
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

// Initializes the external ip address
func initExternal() net.IP {
	external, ok := getExternal(
		"http://ipv4.icanhazip.com",
		"http://myexternalip.com/raw",
		"http://www.trackip.net/ip",
		"http://automation.whatismyip.com/n09230945.asp",
		"http://api.ipify.org/",
	)

	// Make sure we have an IP address, otherwise panic
	if !ok || external == nil {
		logging.LogAction("address", "unable to retrieve external IP address")
		return net.ParseIP("127.0.0.1")
	}

	return external
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

// getPrivateAddr gets the first private address
func getPrivateAddr() (addr string) {
	addr = "127.0.0.1"
	if privateAddrs, err := sockaddr.GetPrivateIPs(); err == nil {
		addrs := strings.Split(privateAddrs, " ")
		if len(addrs) > 0 {
			addr = addrs[0]
		}
	}
	return
}

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
		addr = strings.Replace(addr, "private", getPrivateAddr(), 1)
	}

	// Convenience: set public address
	if strings.Contains(addr, "public") {
		addr = strings.Replace(addr, "public", External().String(), 1)
	}

	// If we have only an IP address, use the default port
	if ip := net.ParseIP(addr); ip != nil {
		addr = fmt.Sprintf("%s:%d", ip, defaultPort)
	}

	// Resolve the address
	return net.ResolveTCPAddr("tcp", addr)
}
