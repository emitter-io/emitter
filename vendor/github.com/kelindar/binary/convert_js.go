// +build js
// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

func binaryToString(buf *[]byte) string {
	return string(*buf)
}

func stringToBinary(v string) (b []byte) {
	return []byte(v)
}

func binaryToBools(b *[]byte) (result []bool) {
	buffer := *b
	result = make([]bool, len(buffer), len(buffer))
	for i := 0; i < len(buffer); i++ {
		result[i] = buffer[i] == 1
	}
	return
}

func boolsToBinary(v *[]bool) (result []byte) {
	value := *v
	result = make([]byte, len(value), len(value))
	for i := 0; i < len(value); i++ {
		if value[i] {
			result[i] = 1
		}
	}
	return
}
