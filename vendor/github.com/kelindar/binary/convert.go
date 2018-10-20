// +build !js
// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

import (
	"reflect"
	"unsafe"
)

func binaryToString(b *[]byte) string {
	return *(*string)(unsafe.Pointer(b))
}

func stringToBinary(v string) (b []byte) {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&v))
	byteHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	byteHeader.Data = strHeader.Data

	l := len(v)
	byteHeader.Len = l
	byteHeader.Cap = l
	return
}

func binaryToBools(b *[]byte) []bool {
	return *(*[]bool)(unsafe.Pointer(b))
}

func boolsToBinary(v *[]bool) []byte {
	return *(*[]byte)(unsafe.Pointer(v))
}
