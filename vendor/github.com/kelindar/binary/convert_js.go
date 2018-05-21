// +build js
// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

func convertToString(buf *[]byte) string {
	return string(*buf)
}

func convertToBytes(v string) (b []byte) {
	return []byte(v)
}
