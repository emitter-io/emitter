package binary

import (
	"encoding/binary"
	"errors"
	"reflect"
	"unsafe"
)

// Constants
var (
	LittleEndian  = binary.LittleEndian
	BigEndian     = binary.BigEndian
	DefaultEndian = LittleEndian
)

// Codec represents a single part codec, which can encode something.
type codec interface {
	EncodeTo(*Encoder, reflect.Value) error
	DecodeTo(*Decoder, reflect.Value) error
}

// ------------------------------------------------------------------------------

type reflectArrayCodec struct {
	elemCodec codec // The codec of the array's elements
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
	elemCodec codec // The codec of the slice's elements
}

// Encode encodes a value into the encoder.
func (c *reflectSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.writeUint64(uint64(l))
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
	l := rv.Len()
	e.writeUint64(uint64(l))
	e.write(rv.Bytes())
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *byteSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil && l > 0 {
		buffer := make([]byte, int(l))
		if _, err = d.r.Read(buffer); err == nil {
			rv.Set(reflect.ValueOf(buffer))
		}
	}
	return
}

// ------------------------------------------------------------------------------

type varintSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *varintSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.writeUint64(uint64(l))
	for i := 0; i < l; i++ {
		e.writeInt64(rv.Index(i).Int())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *varintSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil {
		rv.Set(reflect.MakeSlice(rv.Type(), int(l), int(l)))
		for i := 0; i < int(l); i++ {
			var v int64
			if v, err = binary.ReadVarint(d.r); err == nil {
				reflect.Indirect(rv.Index(i)).SetInt(v)
			}
		}
	}
	return
}

// ------------------------------------------------------------------------------

type varuintSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *varuintSliceCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.writeUint64(uint64(l))
	for i := 0; i < l; i++ {
		e.writeUint64(rv.Index(i).Uint())
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *varuintSliceCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil {
		rv.Set(reflect.MakeSlice(rv.Type(), int(l), int(l)))
		for i := 0; i < int(l); i++ {
			var v uint64
			if v, err = binary.ReadUvarint(d.r); err == nil {
				reflect.Indirect(rv.Index(i)).SetUint(v)
			}
		}
	}
	return
}

// ------------------------------------------------------------------------------

type reflectStructCodec struct {
	fields []fieldCodec // Codecs for all of the fields of the struct
}

type fieldCodec struct {
	index int
	codec codec
}

// Encode encodes a value into the encoder.
func (c *reflectStructCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	for _, i := range c.fields {
		if err = i.codec.EncodeTo(e, rv.Field(i.index)); err != nil {
			return
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectStructCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	for _, i := range c.fields {
		if v := rv.Field(i.index); v.CanSet() {
			if err = i.codec.DecodeTo(d, reflect.Indirect(v.Addr())); err != nil {
				return
			}
		}
	}
	return
}

// ------------------------------------------------------------------------------

type customMarshalCodec struct {
	marshaler      *reflect.Method
	unmarshaler    *reflect.Method
	ptrMarshaler   *reflect.Method
	ptrUnmarshaler *reflect.Method
}

// Encode encodes a value into the encoder.
func (c *customMarshalCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
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
	e.writeUint64(uint64(len(buffer)))
	e.write(buffer)
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *customMarshalCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
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

func (c *customMarshalCodec) GetMarshalBinary(rv reflect.Value) *reflect.Value {
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

func (c *customMarshalCodec) GetUnmarshalBinary(rv reflect.Value) *reflect.Value {
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
	key codec // Codec for the key
	val codec // Codec for the value
}

// Encode encodes a value into the encoder.
func (c *reflectMapCodec) EncodeTo(e *Encoder, rv reflect.Value) (err error) {
	e.writeUint64(uint64(rv.Len()))

	for _, key := range rv.MapKeys() {
		value := rv.MapIndex(key)
		if err = c.key.EncodeTo(e, key); err != nil {
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
	if l, err = binary.ReadUvarint(d.r); err != nil {
		return
	}

	t := rv.Type()
	kt := t.Key()
	vt := t.Elem()
	rv.Set(reflect.MakeMap(t))
	for i := 0; i < int(l); i++ {
		kv := reflect.Indirect(reflect.New(kt))
		if err = c.key.DecodeTo(d, kv); err != nil {
			return
		}

		vv := reflect.Indirect(reflect.New(vt))
		if err = c.val.DecodeTo(d, vv); err != nil {
			return
		}

		rv.SetMapIndex(kv, vv)
	}
	return
}

// ------------------------------------------------------------------------------

type stringCodec struct{}

// Encode encodes a value into the encoder.
func (c *stringCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeString(rv.String())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *stringCodec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = binary.ReadUvarint(d.r); err == nil {
		buf := make([]byte, l)
		_, err = d.r.Read(buf)

		rv.SetString(*(*string)(unsafe.Pointer(&buf)))
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
	var out byte
	err = binary.Read(d.r, d.Order, &out)
	rv.SetBool(out != 0)
	return
}

// ------------------------------------------------------------------------------

type varintCodec struct{}

// Encode encodes a value into the encoder.
func (c *varintCodec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeInt64(rv.Int())
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
	e.writeUint64(rv.Uint())
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
	e.writeComplex(rv.Complex())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *complex64Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out complex64
	err = binary.Read(d.r, d.Order, &out)
	rv.SetComplex(complex128(out))
	return
}

// ------------------------------------------------------------------------------

type complex128Codec struct{}

// Encode encodes a value into the encoder.
func (c *complex128Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeComplex(rv.Complex())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *complex128Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out complex128
	err = binary.Read(d.r, d.Order, &out)
	rv.SetComplex(out)
	return
}

// ------------------------------------------------------------------------------

type float32Codec struct{}

// Encode encodes a value into the encoder.
func (c *float32Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeFloat(rv.Float())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float32Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out float32
	err = binary.Read(d.r, d.Order, &out)
	rv.SetFloat(float64(out))
	return
}

// ------------------------------------------------------------------------------

type float64Codec struct{}

// Encode encodes a value into the encoder.
func (c *float64Codec) EncodeTo(e *Encoder, rv reflect.Value) error {
	e.writeFloat(rv.Float())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float64Codec) DecodeTo(d *Decoder, rv reflect.Value) (err error) {
	var out float64
	err = binary.Read(d.r, d.Order, &out)
	rv.SetFloat(out)
	return
}
