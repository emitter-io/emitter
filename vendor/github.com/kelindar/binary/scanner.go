// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

import (
	"errors"
	"reflect"
	"sync"
)

// Map of all the schemas we've encountered so far
var schemas = new(sync.Map)

// Scan gets a codec for the type and uses a cached schema if the type was
// previously scanned.
func scan(t reflect.Type) (c codec, err error) {

	// Attempt to load from cache first
	if f, ok := schemas.Load(t); ok {
		c = f.(codec)
		return
	}

	// Scan for the first time
	c, err = scanType(t)
	if err != nil {
		return
	}

	// Load or store again
	if f, ok := schemas.LoadOrStore(t, c); ok {
		c = f.(codec)
		return
	}
	return
}

// ScanType scans the type
func scanType(t reflect.Type) (codec, error) {
	if custom, ok := scanCustomCodec(t); ok {
		return custom, nil
	}

	switch t.Kind() {
	case reflect.Array:
		elemCodec, err := scanType(t.Elem())
		if err != nil {
			return nil, err
		}

		return &reflectArrayCodec{
			elemCodec: elemCodec,
		}, nil

	case reflect.Slice:

		// Fast-paths for simple numeric slices and string slices
		switch t.Elem().Kind() {
		case reflect.Int8:
			fallthrough
		case reflect.Uint8:
			return new(byteSliceCodec), nil

		case reflect.Uint:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			return new(varuintSliceCodec), nil

		case reflect.Int:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			return new(varintSliceCodec), nil

		default:
			elemCodec, err := scanType(t.Elem())
			if err != nil {
				return nil, err
			}

			return &reflectSliceCodec{
				elemCodec: elemCodec,
			}, nil
		}

	case reflect.Struct:
		s := scanStruct(t)
		var v reflectStructCodec
		for _, i := range s.fields {
			if c, err := scanType(t.Field(i).Type); err == nil {
				v.fields = append(v.fields, fieldCodec{index: i, codec: c})
			} else {
				return nil, err
			}
		}

		return &v, nil

	case reflect.Map:
		key, err := scanType(t.Key())
		if err != nil {
			return nil, err
		}

		val, err := scanType(t.Elem())
		if err != nil {
			return nil, err
		}

		return &reflectMapCodec{
			key: key,
			val: val,
		}, nil

	case reflect.String:
		return new(stringCodec), nil

	case reflect.Bool:
		return new(boolCodec), nil

	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int64:
		return new(varintCodec), nil

	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint64:
		return new(varuintCodec), nil

	case reflect.Complex64:
		return new(complex64Codec), nil

	case reflect.Complex128:
		return new(complex128Codec), nil

	case reflect.Float32:
		return new(float32Codec), nil

	case reflect.Float64:
		return new(float64Codec), nil
	}

	return nil, errors.New("binary: unsupported type " + t.String())
}

type scannedStruct struct {
	fields []int
}

func scanStruct(t reflect.Type) (meta *scannedStruct) {
	l := t.NumField()
	meta = new(scannedStruct)
	for i := 0; i < l; i++ {
		if t.Field(i).Name != "_" {
			meta.fields = append(meta.fields, i)
		}
	}
	return
}

// ScanCustom scans whether a type has a custom marshaling implemented.
func scanCustomCodec(t reflect.Type) (out *customCodec, ok bool) {
	out = new(customCodec)
	if m, ok := t.MethodByName("MarshalBinary"); ok {
		out.marshaler = &m
	} else if m, ok := reflect.PtrTo(t).MethodByName("MarshalBinary"); ok {
		out.ptrMarshaler = &m
	}

	if m, ok := t.MethodByName("UnmarshalBinary"); ok {
		out.unmarshaler = &m
	} else if m, ok := reflect.PtrTo(t).MethodByName("UnmarshalBinary"); ok {
		out.ptrUnmarshaler = &m
	}

	// Checks whether we have both marshaler and unmarshaler attached
	if (out.marshaler != nil || out.ptrMarshaler != nil) && (out.unmarshaler != nil || out.ptrUnmarshaler != nil) {
		return out, true
	}

	return nil, false
}
