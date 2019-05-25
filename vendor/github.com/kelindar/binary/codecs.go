// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

import (
	"encoding/binary"
	"errors"
	"reflect"
)

// Constants
var (
	LittleEndian = binary.LittleEndian
	BigEndian    = binary.BigEndian
)

// Codec represents a single part Codec, which can encode and decode something.
type Codec interface {
	EncodeTo(*Encoder, reflect.Value) error
	DecodeTo(*Decoder, reflect.Value) error
}

// ------------------------------------------------------------------------------

type reflectArrayCodec struct {
	elemCodec Codec // The codec of the array's elements
}

// Encode encodes a value into the encoder.
func (c *reflectArrayCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Type().Len()
	for i := 0; i < l; i++ {
		v := reflect.Indirect(rv.Index(i).Addr())
		if err = c.elemCodec.EncodeTo(e, v); err != nil {
			return
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectArrayCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	l := rv.Type().Len()
	for i := 0; i < l; i++ {
		v := reflect.Indirect(rv.Index(i))
		if err = c.elemCodec.DecodeTo(d, v); err != nil {
			return
		}
	}
	return
}

// ------------------------------------------------------------------------------

type reflectSliceCodec struct {
	elemCodec Codec // The codec of the slice's elements
}

// Encode encodes a value into the encoder.
func (c *reflectSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		v := reflect.Indirect(rv.Index(i).Addr())
		if err = c.elemCodec.EncodeTo(e, v); err != nil {
			return
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil && l > 0 {
		rv.Set(reflect.MakeSlice(rv.Type(), int(l), int(l)))
		for i := 0; i < int(l); i++ {
			v := reflect.Indirect(rv.Index(i))
			if err = c.elemCodec.DecodeTo(d, v); err != nil {
				return
			}
		}
	}
	return
}

// ------------------------------------------------------------------------------

type byteSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *byteSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	b := rv.Bytes()
	e.WriteUvarint(uint64(len(b)))
	e.Write(b)
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *byteSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		data := make([]byte, int(l), int(l))
		if _, err = d.Read(data); err == nil {
			rv.Set(reflect.ValueOf(data))
		}
	}
	return
}

// ------------------------------------------------------------------------------

type boolSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *boolSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.WriteUvarint(uint64(l))
	if l > 0 {
		v := rv.Interface().([]bool)
		e.Write(boolsToBinary(&v))
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *boolSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		buf := make([]byte, l)
		_, err = d.r.Read(buf)
		rv.Set(reflect.ValueOf(binaryToBools(&buf)))
	}
	return
}

// ------------------------------------------------------------------------------

type varintSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *varintSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		e.WriteVarint(rv.Index(i).Int())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *varintSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil && l > 0 {
		slice := reflect.MakeSlice(rv.Type(), int(l), int(l))
		for i := 0; i < int(l); i++ {
			var v int64
			if v, err = binary.ReadVarint(d.r); err == nil {
				slice.Index(i).SetInt(v)
			}
		}

		rv.Set(slice)
	}
	return
}

// ------------------------------------------------------------------------------

type varuintSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *varuintSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		e.WriteUvarint(rv.Index(i).Uint())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *varuintSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l, v uint64
	if l, err = binary.ReadUvarint(d.r); err == nil && l > 0 {
		slice := reflect.MakeSlice(rv.Type(), int(l), int(l))
		for i := 0; i < int(l); i++ {
			if v, err = d.ReadUvarint(); err == nil {
				slice.Index(i).SetUint(v)
			}
		}

		rv.Set(slice)
	}
	return
}

// ------------------------------------------------------------------------------

type reflectStructCodec []fieldCodec

type fieldCodec struct {
	Index int   // The index of the field
	Codec Codec // The codec to use for this field
}

// Encode encodes a value into the encoder.
func (c *reflectStructCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	for _, i := range *c {
		if err = i.Codec.EncodeTo(e, rv.Field(i.Index)); err != nil {
			return
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectStructCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	for _, i := range *c {
		if v := rv.Field(i.Index); v.CanSet() {
			if err = i.Codec.DecodeTo(d, reflect.Indirect(v)); err != nil {
				return
			}
		}
	}
	return
}

// ------------------------------------------------------------------------------

// customCodec represents a custom binary marshaling.
type customCodec struct {
	marshaler      *reflect.Method
	unmarshaler    *reflect.Method
	ptrMarshaler   *reflect.Method
	ptrUnmarshaler *reflect.Method
}

// Encode encodes a value into the encoder.
func (c *customCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	m := c.GetMarshalBinary(rv)
	if m == nil {
		return errors.New("MarshalBinary not found on " + rv.Type().String())
	}

	ret := m.Call([]reflect.Value{})
	if !ret[1].IsNil() {
		err = ret[1].Interface().(error)
		return
	}

	// Write the marshaled byte slice
	buffer := ret[0].Bytes()
	e.WriteUvarint(uint64(len(buffer)))
	e.Write(buffer)
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *customCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	m := c.GetUnmarshalBinary(rv)

	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil {
		buffer := make([]byte, l)
		_, err = d.r.Read(buffer)
		ret := m.Call([]reflect.Value{reflect.ValueOf(buffer)})
		if !ret[0].IsNil() {
			err = ret[0].Interface().(error)
		}

	}
	return
}

func (c *customCodec) GetMarshalBinary(rv reflect.Value) *reflect.Value {
	if c.marshaler != nil {
		m := rv.Method(c.marshaler.Index)
		return &m
	}

	if c.ptrMarshaler != nil {
		m := rv.Addr().Method(c.ptrMarshaler.Index)
		return &m
	}

	return nil
}

func (c *customCodec) GetUnmarshalBinary(rv reflect.Value) *reflect.Value {
	if c.unmarshaler != nil {
		m := rv.Method(c.unmarshaler.Index)
		return &m
	}

	if c.ptrUnmarshaler != nil {
		m := rv.Addr().Method(c.ptrUnmarshaler.Index)
		return &m
	}

	return nil
}

// ------------------------------------------------------------------------------

type reflectMapCodec struct {
	key Codec // Codec for the key
	val Codec // Codec for the value
}

// Encode encodes a value into the encoder.
func (c *reflectMapCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	e.WriteUvarint(uint64(rv.Len()))
	for _, key := range rv.MapKeys() {
		value := rv.MapIndex(key)
		if err = c.writeKey(e, key); err != nil {
			return err
		}

		if err = c.val.EncodeTo(e, value); err != nil {
			return err
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectMapCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil {
		t := rv.Type()
		vt := t.Elem()
		rv.Set(reflect.MakeMap(t))
		for i := 0; i < int(l); i++ {

			var kv reflect.Value
			if kv, err = c.readKey(d, t.Key()); err != nil {
				return
			}

			vv := reflect.Indirect(reflect.New(vt))
			if err = c.val.DecodeTo(d, vv); err != nil {
				return
			}

			rv.SetMapIndex(kv, vv)
		}
	}
	return
}

// Write key writes a key to the encoder
func (c *reflectMapCodec) writeKey(e *Encoder, key reflect.Value) (err error) {
	switch key.Kind() {

	case reflect.Int16:
		e.WriteUint16(uint16(key.Int()))
	case reflect.Int32:
		e.WriteUint32(uint32(key.Int()))
	case reflect.Int64:
		e.WriteUint64(uint64(key.Int()))

	case reflect.Uint16:
		e.WriteUint16(uint16(key.Uint()))
	case reflect.Uint32:
		e.WriteUint32(uint32(key.Uint()))
	case reflect.Uint64:
		e.WriteUint64(uint64(key.Uint()))

	case reflect.String:
		str := key.String()
		e.WriteUint16(uint16(len(str)))
		e.Write(stringToBinary(str))
	default:
		err = c.key.EncodeTo(e, key)
	}
	return
}

// Read key reads a key from the decoder
func (c *reflectMapCodec) readKey(d *Decoder, keyType reflect.Type) (key reflect.Value, err error) {
	switch keyType.Kind() {

	case reflect.Int16:
		var v uint16
		if v, err = d.ReadUint16(); err == nil {
			key = reflect.ValueOf(int16(v))
		}
	case reflect.Int32:
		var v uint32
		if v, err = d.ReadUint32(); err == nil {
			key = reflect.ValueOf(int32(v))
		}
	case reflect.Int64:
		var v uint64
		if v, err = d.ReadUint64(); err == nil {
			key = reflect.ValueOf(int64(v))
		}

	case reflect.Uint16:
		var v uint16
		if v, err = d.ReadUint16(); err == nil {
			key = reflect.ValueOf(v)
		}
	case reflect.Uint32:
		var v uint32
		if v, err = d.ReadUint32(); err == nil {
			key = reflect.ValueOf(v)
		}
	case reflect.Uint64:
		var v uint64
		if v, err = d.ReadUint64(); err == nil {
			key = reflect.ValueOf(v)
		}

	// String keys must have max length of 65536
	case reflect.String:
		var l uint16
		var b []byte

		if l, err = d.ReadUint16(); err == nil {
			if b, err = d.Slice(int(l)); err == nil {
				key = reflect.ValueOf(string(b))
			}
		}

	// Default to a reflect-based approach
	default:
		key = reflect.Indirect(reflect.New(keyType))
		err = c.key.DecodeTo(d, key)
	}
	return
}

// ------------------------------------------------------------------------------

type stringCodec struct{}

// Encode encodes a value into the encoder.
func (c *stringCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	str := rv.String()
	e.WriteUvarint(uint64(len(str)))
	e.Write(stringToBinary(str))
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *stringCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	var b []byte

	if l, err = d.ReadUvarint(); err == nil {
		if b, err = d.Slice(int(l)); err == nil {
			rv.SetString(string(b))
		}
	}
	return
}

// ------------------------------------------------------------------------------

type boolCodec struct{}

// Encode encodes a value into the encoder.
func (c *boolCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeBool(rv.Bool())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *boolCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out bool
	if out, err = d.ReadBool(); err == nil {
		rv.SetBool(out)
	}
	return
}

// ------------------------------------------------------------------------------

type varintCodec struct{}

// Encode encodes a value into the encoder.
func (c *varintCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.WriteVarint(rv.Int())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *varintCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var v int64
	if v, err = binary.ReadVarint(d.r); err != nil {
		return
	}
	rv.SetInt(v)
	return
}

// ------------------------------------------------------------------------------

type varuintCodec struct{}

// Encode encodes a value into the encoder.
func (c *varuintCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.WriteUvarint(rv.Uint())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *varuintCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var v uint64
	if v, err = binary.ReadUvarint(d.r); err != nil {
		return
	}
	rv.SetUint(v)
	return
}

// ------------------------------------------------------------------------------

type complex64Codec struct{}

// Encode encodes a value into the encoder.
func (c *complex64Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeComplex64(complex64(rv.Complex()))
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *complex64Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out complex64
	out, err = d.readComplex64()
	rv.SetComplex(complex128(out))
	return
}

// ------------------------------------------------------------------------------

type complex128Codec struct{}

// Encode encodes a value into the encoder.
func (c *complex128Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeComplex128(rv.Complex())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *complex128Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out complex128
	out, err = d.readComplex128()
	rv.SetComplex(out)
	return
}

// ------------------------------------------------------------------------------

type float32Codec struct{}

// Encode encodes a value into the encoder.
func (c *float32Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.WriteFloat32(float32(rv.Float()))
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float32Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var v float32
	if v, err = d.ReadFloat32(); err == nil {
		rv.SetFloat(float64(v))
	}
	return
}

// ------------------------------------------------------------------------------

type float64Codec struct{}

// Encode encodes a value into the encoder.
func (c *float64Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.WriteFloat64(rv.Float())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float64Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var v float64
	if v, err = d.ReadFloat64(); err == nil {
		rv.SetFloat(v)
	}
	return
}
