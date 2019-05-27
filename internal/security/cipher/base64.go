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

package cipher

import "strconv"

// The map used for base64 decode.
var decodeMap [256]byte

type corruptInputError int64

func (e corruptInputError) Error() string {
	return "illegal base64 data at input byte " + strconv.FormatInt(int64(e), 10)
}

// Init prepares the lookup table.
func init() {
	encoder := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	for i := 0; i < len(decodeMap); i++ {
		decodeMap[i] = 0xFF
	}
	for i := 0; i < len(encoder); i++ {
		decodeMap[encoder[i]] = byte(i)
	}
}

// decodeKey decodes the key from base64 string, url-encoded with no
// padding. This is 2x faster than the built-in function as we trimmed
// it significantly.
func decodeKey(dst, src []byte) (n int, err error) {
	var idx int
	for idx < len(src) {
		var dbuf [4]byte
		dinc, dlen := 3, 4

		for j := range dbuf {
			if len(src) == idx {
				if j < 2 {
					return n, corruptInputError(idx - j)
				}
				dinc, dlen = j-1, j
				break
			}

			in := src[idx]
			idx++

			dbuf[j] = decodeMap[in]
			if dbuf[j] == 0xFF {
				return n, corruptInputError(idx - 1)
			}
		}

		// Convert 4x 6bit source bytes into 3 bytes
		val := uint(dbuf[0])<<18 | uint(dbuf[1])<<12 | uint(dbuf[2])<<6 | uint(dbuf[3])
		dbuf[2], dbuf[1], dbuf[0] = byte(val>>0), byte(val>>8), byte(val>>16)
		switch dlen {
		case 4:
			dst[2] = dbuf[2]
			dbuf[2] = 0
			fallthrough
		case 3:
			dst[1] = dbuf[1]
			dbuf[1] = 0
			fallthrough
		case 2:
			dst[0] = dbuf[0]
		}

		dst = dst[dinc:]
		n += dlen - 1
	}

	return n, err
}
