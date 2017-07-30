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
	"encoding/base64"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

var fingerprint string
var external net.IP

func init() {
	external = initExternal()
	fingerprint = initFingerprint()
}

// External gets the external IP address
func External() net.IP {
	return external
}

// Fingerprint gets the fingerprint
func Fingerprint() string {
	return fingerprint
}

// getExternal retrieves an external IP address
func getExternal(urls ...string) (net.IP, bool) {
	for _, url := range urls {
		cli := http.Client{Timeout: time.Duration(5 * time.Second)}
		res, err := cli.Get(url)
		if err != nil {
			continue
		}

		// Read the response
		defer res.Body.Close()
		r, err := ioutil.ReadAll(res.Body)
		if err != nil {
			continue
		}

		// Fix and parse
		addr := strings.Replace(string(r), "\n", "", -1)
		ip := net.ParseIP(addr)
		return ip, ip != nil
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
		panic("Unable to retrieve external IP address")
	}
	return external
}

// Initializes the fingerprint
func initFingerprint() string {
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

	// Initialize hardwareAddr randomly in case of real network interfaces absence
	safeRandom(hardwareAddr[:])

	// Set multicast bit as recommended in RFC 4122
	hardwareAddr[0] |= 0x01

	return encode(hardwareAddr[:])
}

func encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func safeRandom(dest []byte) {
	if _, err := rand.Read(dest); err != nil {
		panic(err)
	}
}
