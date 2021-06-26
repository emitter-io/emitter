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
	"fmt"
	"strconv"
	"time"

	"github.com/emitter-io/emitter/internal/config"
	"github.com/emitter-io/emitter/internal/security/hash"
	"github.com/kelindar/binary"
)

// Channel types
const (
	ChannelInvalid = uint8(iota)
	ChannelStatic
	ChannelWildcard
)

// Minimum and maximum unix time stamp we can handle for options
const (
	MinTime = 1514764800 // 2018
	MaxTime = 3029529600 // 2066
)

var zeroTime = time.Unix(0, 0)

// ChannelOption represents a key/value pair option.
type ChannelOption struct {
	Key   string
	Value string
}

// Channel represents a parsed MQTT topic.
type Channel struct {
	Key         []byte          // Gets or sets the API key of the channel.
	Channel     []byte          // Gets or sets the channel string.
	Query       []uint32        // Gets or sets the full ssid.
	Options     []ChannelOption // Gets or sets the options.
	ChannelType uint8
}

// Target returns the channel target (first element of the query, second element of an SSID)
func (c *Channel) Target() uint32 {
	return c.Query[0]
}

// TTL returns a Time-To-Live option.
func (c *Channel) TTL() (int64, bool) {
	return c.getOption("ttl", 64)
}

// Last returns the 'last' option, which is a number of messages to retrieve.
func (c *Channel) Last() (int64, bool) {
	return c.getOption("last", 64)
}

// Exclude returns whether the exclude me ('me=0') option was set or not.
func (c *Channel) Exclude() bool {
	v, ok := c.getOption("me", 64)
	return ok && v == 0
}

// Window returns the from-until options which should be a UTC unix timestamp in seconds.
func (c *Channel) Window() (time.Time, time.Time) {
	u0, _ := c.getOption("from", 64)
	u1, _ := c.getOption("until", 64)
	return toUnix(u0), toUnix(u1)
}

// SafeString returns a string representation of the channel without the key.
func (c *Channel) SafeString() string {
	text := string(c.Channel)
	if len(c.Options) == 0 {
		return text
	}

	text += "?"
	for i, v := range c.Options {
		if i > 0 {
			text += "&"
		}

		text += v.Key + "=" + v.Value
	}
	return text
}

// String returns a string representation of the channel.
func (c *Channel) String() string {
	text := string(c.Key)
	text += "/"
	text += c.SafeString()
	return text
}

// Converts the time to Unix Time with validation.
func toUnix(t int64) time.Time {
	if t == 0 || t < MinTime || t > MaxTime {
		return zeroTime
	}

	return time.Unix(t, 0)
}

// getOptUint retrieves a Uint option
func (c *Channel) getOption(name string, bitSize int) (int64, bool) {
	for i := 0; i < len(c.Options); i++ {
		if c.Options[i].Key == name {
			if val, err := strconv.ParseInt(c.Options[i].Value, 10, bitSize); err == nil {
				return int64(val), true
			}
			return 0, false
		}
	}
	return 0, false
}

// MakeChannel attempts to parse the channel from the key and channel strings.
func MakeChannel(key, channelWithOptions string) *Channel {
	return ParseChannel([]byte(fmt.Sprintf("%s/%s", key, channelWithOptions)))
}

// ParseChannel attempts to parse the channel from the underlying slice.
func ParseChannel(text []byte) (channel *Channel) {
	channel = new(Channel)
	channel.Query = make([]uint32, 0, 6)
	offset := 0

	// First we need to parse the key part
	i, ok := channel.parseKey(text)
	if !ok {
		channel.ChannelType = ChannelInvalid
		return channel
	}

	// Now parse the channel
	offset += i
	i = channel.parseChannel(text[offset:])
	if channel.ChannelType == ChannelInvalid {
		return channel
	}

	// Now parse the options
	offset += i
	if offset < len(text) {
		_, ok = channel.parseOptions(text[offset:])
		if !ok {
			channel.ChannelType = ChannelInvalid
			return channel
		}
	}

	// We've processed everything now
	return channel
}

// ParseKey reads the provided API key, this should be the 32-character long
// key or 'emitter' string for custom API requests.
func (c *Channel) parseKey(text []byte) (i int, ok bool) {
	//keyChars := 0
	for ; i < len(text); i++ {
		if text[i] == config.ChannelSeparator {
			if c.Key = text[:i]; len(c.Key) > 0 {
				return i + 1, true
			}
			break
		}
	}
	return i, false
}

// ParseKey reads the provided API key, this should be the 32-character long
// key or 'emitter' string for custom API requests.
func (c *Channel) parseChannel(text []byte) (i int) {
	length, offset := len(text), 0
	chanChars := 0
	wildcards := 0
	for ; i < length; i++ {
		symbol := text[i] // The current byte
		switch {

		// If we're reading a separator compute the SSID.
		case symbol == config.ChannelSeparator:
			if chanChars == 0 && wildcards == 0 {
				c.ChannelType = ChannelInvalid
				return i
			}
			c.Query = append(c.Query, hash.Of(text[offset:i]))

			if i+1 == length { // The end flag
				c.Channel = text[:i+1]
				if c.ChannelType != ChannelWildcard {
					c.ChannelType = ChannelStatic
				}
				return i + 1
			} else if text[i+1] == '?' {
				c.Channel = text[:i+1]
				if c.ChannelType != ChannelWildcard {
					c.ChannelType = ChannelStatic
				}
				return i + 2
			}

			offset = i + 1
			chanChars = 0
			wildcards = 0
			continue
		// If this symbol is a wildcard symbol
		case symbol == '#' || symbol == '+' || symbol == '*':
			if chanChars > 0 || wildcards > 0 {
				c.ChannelType = ChannelInvalid
				return i
			}
			wildcards++
			c.ChannelType = ChannelWildcard
			continue

		// Valid character, but nothing special
		case (symbol >= 45 && symbol <= 58) || (symbol >= 65 && symbol <= 122) || symbol == 36:
			if wildcards > 0 {
				c.ChannelType = ChannelInvalid
				return i
			}
			chanChars++
			continue

		// Weird character, fail.
		default:
			c.ChannelType = ChannelInvalid
			return i
		}
	}
	c.ChannelType = ChannelInvalid
	return i
}

// ParseOptions parses the key/value pairs of options, encoded as URL Query string.
func (c *Channel) parseOptions(text []byte) (i int, ok bool) {
	length := len(text)
	j := i

	// We need to create the options container, if we do have options
	c.Options = make([]ChannelOption, 0, 2)
	var key, val []byte

	//chanChars := 0

	// Start reading the options.
	for i < length {

		// Get the key
		for j < length {
			symbol := text[j] // The current byte
			j++

			if symbol == '=' {
				key = text[i : j-1]
				i = j
				break
			} else if !((symbol >= 48 && symbol <= 57) || (symbol >= 65 && symbol <= 90) || (symbol >= 97 && symbol <= 122)) {
				return i, false
			}
		}

		// Get the value
		for j < length {
			symbol := text[j]
			j++

			if symbol == '&' {
				val = text[i : j-1]
				i = j
				break
			} else if !((symbol >= 48 && symbol <= 57) || (symbol >= 65 && symbol <= 90) || (symbol >= 97 && symbol <= 122)) {
				return i, false
			} else if j == length {
				val = text[i:j]
				i = j
				// break ? and what about goto for perfs ?
			}
		}

		// By now we should have a key and a value, otherwise this is not a valid channel string.
		if len(key) == 0 || len(val) == 0 {
			return i, false
		}

		// Set the option
		c.Options = append(c.Options, ChannelOption{
			Key:   binary.ToString(&key),
			Value: binary.ToString(&val),
		})

		val = val[0:0]
		key = key[0:0]
	}

	return i, true
}
