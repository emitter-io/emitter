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
	"regexp"
)

// This is a strict format for the dial string
var dialFormat = regexp.MustCompile(`^(dial)\:\/\/(.+)$`)

// ParseDial parses a pre-authorized channel key
func ParseDial(password string) (*Channel, bool) {
	parts := dialFormat.FindStringSubmatch(password)
	if len(parts) != 3 {
		return nil, false // Invalid channel
	}

	// Get the scheme and channel and make sure they're valid
	scheme := parts[1]
	channel := ParseChannel([]byte(parts[2]))
	if len(scheme) == 0 || channel == nil || channel.ChannelType == ChannelInvalid {
		return nil, false
	}

	// For dial to work, the channel must be static
	if scheme == "dial" && channel.ChannelType == ChannelStatic {
		return channel, true
	}

	// Safe default
	return nil, false
}
