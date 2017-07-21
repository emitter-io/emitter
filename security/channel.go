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
	"reflect"
	"strconv"
	"unsafe"

	"github.com/emitter-io/emitter/config"
	"github.com/emitter-io/emitter/utils"
)

// Channel types
const (
	ChannelInvalid = uint8(iota)
	ChannelStatic
	ChannelWildcard
)

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

func (c *Channel) Target() uint32 {
	return c.Query[0]
}

// TODO why not a map of options?
func (c *Channel) TTL() (bool, uint32) {
	for i := 0; i < len(c.Options); i++ {
		if len(c.Options[i].Key) == 3 && c.Options[i].Key == "ttl" {
			if val, err := strconv.ParseUint(c.Options[i].Value, 10, 32); err == nil {
				return true, uint32(val)
			}
			return false, 0
		}
	}
	return false, 0
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
		i, ok = channel.parseOptions(text[offset:])
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
			c.Query = append(c.Query, utils.GetHash(text[offset:i]))

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
		case symbol == '+' || symbol == '*':
			if chanChars > 0 || wildcards > 0 {
				c.ChannelType = ChannelInvalid
				return i
			}
			wildcards++
			c.ChannelType = ChannelWildcard
			continue

		// Valid character, but nothing special
		case (symbol >= 45 && symbol <= 58) || (symbol >= 65 && symbol <= 122):
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
			Key:   unsafeToString(key),
			Value: unsafeToString(val),
		})

		val = val[0:0]
		key = key[0:0]
	}

	return i, true
}

func unsafeToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}
