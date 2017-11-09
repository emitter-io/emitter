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

func scanType(t reflect.Type) (codec, error) {
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
		if s.IsCustom() {
			return &customMarshalCodec{
				marshaler:      s.marshaler,
				unmarshaler:    s.unmarshaler,
				ptrMarshaler:   s.ptrMarshaler,
				ptrUnmarshaler: s.ptrUnmarshaler,
			}, nil
		}

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
	fields         []int
	marshaler      *reflect.Method
	unmarshaler    *reflect.Method
	ptrMarshaler   *reflect.Method
	ptrUnmarshaler *reflect.Method
}

func (s *scannedStruct) IsCustom() bool {
	return (s.marshaler != nil || s.ptrMarshaler != nil) && (s.unmarshaler != nil || s.ptrUnmarshaler != nil)
}

func scanStruct(t reflect.Type) (meta *scannedStruct) {
	l := t.NumField()
	meta = new(scannedStruct)
	for i := 0; i < l; i++ {
		if t.Field(i).Name != "_" {
			meta.fields = append(meta.fields, i)
		}
	}

	if m, ok := t.MethodByName("MarshalBinary"); ok {
		meta.marshaler = &m
	} else if m, ok := reflect.PtrTo(t).MethodByName("MarshalBinary"); ok {
		meta.ptrMarshaler = &m
	}

	if m, ok := t.MethodByName("UnmarshalBinary"); ok {
		meta.unmarshaler = &m
	} else if m, ok := reflect.PtrTo(t).MethodByName("UnmarshalBinary"); ok {
		meta.ptrUnmarshaler = &m
	}

	return
}
