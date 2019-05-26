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

package message

import (
	"fmt"
	"testing"

	"github.com/golang/snappy"
	"github.com/stretchr/testify/assert"
)

func TestCodec_Message(t *testing.T) {

	for i := 0; i < 100; i++ {
		t.Run("codec", func(t *testing.T) {
			t.Parallel()
			msg := newTestMessage(Ssid{1, 2, 3}, "a/b/c/", fmt.Sprintf("message number %v", i))
			buffer := msg.Encode()
			//assert.True(t, len(buffer) >= 57)
			//assert.True(t, len(buffer) <= 58)

			// Decode
			output, err := DecodeMessage(buffer)
			assert.NoError(t, err)
			assert.Equal(t, msg, output)
		})
	}
}

func TestCodec_HappyPath(t *testing.T) {
	frame := Frame{
		newTestMessage(Ssid{1, 2, 3}, "a/b/c/", "hello abc"),
		newTestMessage(Ssid{1, 2, 3}, "a/b/", "hello ab"),
	}

	// Encode
	buffer := frame.Encode()
	assert.True(t, len(buffer) >= 65)

	// Decode
	output, err := DecodeFrame(buffer)
	assert.NoError(t, err)
	assert.Equal(t, frame, output)
}

func TestCodec_Corrupt(t *testing.T) {
	_, err := DecodeFrame([]byte{121, 4, 3, 2, 2, 1, 5, 3, 2})
	assert.Equal(t, "snappy: corrupt input", err.Error())
}

func TestCodec_Invalid(t *testing.T) {
	var out []byte
	out = snappy.Encode(out, []byte{121, 4, 3, 2, 2, 1, 5, 3, 2})

	_, err := DecodeFrame(out)
	assert.Equal(t, "EOF", err.Error())
}
