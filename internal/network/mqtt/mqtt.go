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

package mqtt

import (
	"bytes"
	"fmt"
	"io"

	"github.com/emitter-io/emitter/internal/collection"
	"github.com/emitter-io/emitter/internal/config"
)

// buffers are reusable fixed-side buffers for faster encoding.
var buffers = collection.NewBufferPool(config.EncodingBufferSize)

// reserveForHeader reserves the bytes for a header.
var reserveForHeader = []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

//Message is the interface all our packets will be implementing
type Message interface {
	fmt.Stringer

	Type() uint8
	EncodeTo(w io.Writer) (int, error)
}

// MQTT message types
const (
	TypeOfConnect = uint8(iota + 1)
	TypeOfConnack
	TypeOfPublish
	TypeOfPuback
	TypeOfPubrec
	TypeOfPubrel
	TypeOfPubcomp
	TypeOfSubscribe
	TypeOfSuback
	TypeOfUnsubscribe
	TypeOfUnsuback
	TypeOfPingreq
	TypeOfPingresp
	TypeOfDisconnect
)

// StaticHeader as defined in http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html#fixed-header
type StaticHeader struct {
	DUP    bool
	Retain bool
	QOS    uint8
}

// Connect represents an MQTT connect packet.
type Connect struct {
	ProtoName      []byte
	Version        uint8
	UsernameFlag   bool
	PasswordFlag   bool
	WillRetainFlag bool
	WillQOS        uint8
	WillFlag       bool
	CleanSeshFlag  bool
	KeepAlive      uint16
	ClientID       []byte
	WillTopic      []byte
	WillMessage    []byte
	Username       []byte
	Password       []byte
}

// Connack represents an MQTT connack packet.
// 0x00 connection accepted
// 0x01 refused: unacceptable proto version
// 0x02 refused: identifier rejected
// 0x03 refused server unavailiable
// 0x04 bad user or password
// 0x05 not authorized
type Connack struct {
	ReturnCode uint8
}

// Publish represents an MQTT publish packet.
type Publish struct {
	Header    *StaticHeader
	Topic     []byte
	MessageID uint16
	Payload   []byte
}

//Puback is sent for QOS level one to verify the receipt of a publish
//Qoth the spec: "A PUBACK message is sent by a server in response to a PUBLISH message from a publishing client, and by a subscriber in response to a PUBLISH message from the server."
type Puback struct {
	MessageID uint16
}

//Pubrec is for verifying the receipt of a publish
//Qoth the spec:"It is the second message of the QoS level 2 protocol flow. A PUBREC message is sent by the server in response to a PUBLISH message from a publishing client, or by a subscriber in response to a PUBLISH message from the server."
type Pubrec struct {
	MessageID uint16
}

//Pubrel is a response to pubrec from either the client or server.
type Pubrel struct {
	MessageID uint16
	//QOS1
	Header *StaticHeader
}

//Pubcomp is for saying is in response to a pubrel sent by the publisher
//the final member of the QOS2 flow. both sides have said "hey, we did it!"
type Pubcomp struct {
	MessageID uint16
}

//Subscribe tells the server which topics the client would like to subscribe to
type Subscribe struct {
	Header        *StaticHeader
	MessageID     uint16
	Subscriptions []TopicQOSTuple
}

//Suback is to say "hey, you got it buddy. I will send you messages that fit this pattern"
type Suback struct {
	MessageID uint16
	Qos       []uint8
}

//Unsubscribe is the message to send if you don't want to subscribe to a topic anymore
type Unsubscribe struct {
	Header    *StaticHeader
	MessageID uint16
	Topics    []TopicQOSTuple
}

//Unsuback is to unsubscribe as suback is to subscribe
type Unsuback struct {
	MessageID uint16
}

//Pingreq is a keepalive
type Pingreq struct {
}

//Pingresp is for saying "hey, the server is alive"
type Pingresp struct {
}

//Disconnect is to signal you want to cease communications with the server
type Disconnect struct {
}

//TopicQOSTuple is a struct for pairing the Qos and topic together
//for the QOS' pairs in unsubscribe and subscribe
type TopicQOSTuple struct {
	Qos   uint8
	Topic []byte
}

// DecodePacket decodes the packet from the provided reader.
func DecodePacket(rdr io.Reader, maxMessageSize int64) (Message, error) {
	hdr, sizeOf, messageType, err := decodeStaticHeader(rdr)
	if err != nil {
		return nil, err
	}

	// Check for empty packets
	switch messageType {
	case TypeOfPingreq:
		return &Pingreq{}, nil
	case TypeOfPingresp:
		return &Pingresp{}, nil
	case TypeOfDisconnect:
		return &Disconnect{}, nil
	}

	//check to make sure packet isn't above size limit
	if int64(sizeOf) > maxMessageSize {
		return nil, fmt.Errorf("Message size is too large")
	}

	// Now we can decode the buffer. The problem here is that we have to create
	// a new buffer for the body as we're going to simply create slices around it.
	// There's probably a way to use a "read buffer" provided to reduce allocations.
	buffer := make([]byte, sizeOf)
	_, err = io.ReadFull(rdr, buffer)
	if err != nil {
		return nil, err
	}

	// Decode the body
	var msg Message
	switch messageType {
	case TypeOfConnect:
		msg = decodeConnect(buffer, hdr)
	case TypeOfConnack:
		msg = decodeConnack(buffer, hdr)
	case TypeOfPublish:
		msg = decodePublish(buffer, hdr)
	case TypeOfPuback:
		msg = decodePuback(buffer, hdr)
	case TypeOfPubrec:
		msg = decodePubrec(buffer, hdr)
	case TypeOfPubrel:
		msg = decodePubrel(buffer, hdr)
	case TypeOfPubcomp:
		msg = decodePubcomp(buffer, hdr)
	case TypeOfSubscribe:
		msg = decodeSubscribe(buffer, hdr)
	case TypeOfSuback:
		msg = decodeSuback(buffer, hdr)
	case TypeOfUnsubscribe:
		msg = decodeUnsubscribe(buffer, hdr)
	case TypeOfUnsuback:
		msg = decodeUnsuback(buffer, hdr)
	default:
		return nil, fmt.Errorf("Invalid zero-length packet with type %d", messageType)
	}

	return msg, nil
}

// encodeParts sews the whole packet together
func encodeParts(msgType uint8, buf *bytes.Buffer, h *StaticHeader) []byte {
	var firstByte byte
	firstByte |= msgType << 4
	if h != nil {
		firstByte |= boolToUInt8(h.DUP) << 3
		firstByte |= h.QOS << 1
		firstByte |= boolToUInt8(h.Retain)
	}

	// get the length first
	numBytes, bitField := encodeLength(uint32(buf.Len()) - 6)
	offset := 6 - numBytes - 1 //to account for the first byte
	byteBuf := buf.Bytes()

	// now we blit it in
	byteBuf[offset] = byte(firstByte)
	for i := offset + 1; i < 6; i++ {
		//coercing to byte selects the last 8 bits
		byteBuf[i] = byte(bitField >> ((numBytes - 1) * 8))
		numBytes--
	}

	//and return a slice from the offset
	return byteBuf[offset:]
}

// EncodeTo writes the encoded message to the underlying writer.
func (c *Connect) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	//write some buffer space in the beginning for the maximum number of bytes static header + legnth encoding can take
	buf.Write(reserveForHeader)

	// pack the proto name and version
	writeString(buf, c.ProtoName)
	buf.WriteByte(c.Version)

	// pack the flags
	var flagByte byte
	flagByte |= byte(boolToUInt8(c.UsernameFlag)) << 7
	flagByte |= byte(boolToUInt8(c.PasswordFlag)) << 6
	flagByte |= byte(boolToUInt8(c.WillRetainFlag)) << 5
	flagByte |= byte(c.WillQOS) << 3
	flagByte |= byte(boolToUInt8(c.WillFlag)) << 2
	flagByte |= byte(boolToUInt8(c.CleanSeshFlag)) << 1
	buf.WriteByte(flagByte)

	writeUint16(buf, c.KeepAlive)
	writeString(buf, c.ClientID)

	if c.WillFlag {
		writeString(buf, c.WillTopic)
		writeString(buf, c.WillMessage)
	}

	if c.UsernameFlag {
		writeString(buf, c.Username)
	}

	if c.PasswordFlag {
		writeString(buf, c.Password)
	}

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfConnect, buf, nil))
}

// Type returns the MQTT message type.
func (c *Connect) Type() uint8 {
	return TypeOfConnect
}

// String returns the name of mqtt operation.
func (c *Connect) String() string {
	return "connect"
}

// EncodeTo writes the encoded message to the underlying writer.
func (c *Connack) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	//write padding
	buf.Write(reserveForHeader)
	buf.WriteByte(byte(0))
	buf.WriteByte(byte(c.ReturnCode))

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfConnack, buf, nil))
}

// Type returns the MQTT message type.
func (c *Connack) Type() uint8 {
	return TypeOfConnack
}

// String returns the name of mqtt operation.
func (c *Connack) String() string {
	return "connack"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Publish) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	buf.Write(reserveForHeader)
	writeString(buf, p.Topic)
	if p.Header.QOS > 0 {
		writeUint16(buf, p.MessageID)
	}
	buf.Write(p.Payload)

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfPublish, buf, p.Header))
}

// Type returns the MQTT message type.
func (p *Publish) Type() uint8 {
	return TypeOfPublish
}

// String returns the name of mqtt operation.
func (p *Publish) String() string {
	return "pub"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Puback) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	buf.Write(reserveForHeader)
	writeUint16(buf, p.MessageID)

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfPuback, buf, nil))
}

// Type returns the MQTT message type.
func (p *Puback) Type() uint8 {
	return TypeOfPuback
}

// String returns the name of mqtt operation.
func (p *Puback) String() string {
	return "puback"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pubrec) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	buf.Write(reserveForHeader)
	writeUint16(buf, p.MessageID)

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfPubrec, buf, nil))
}

// Type returns the MQTT message type.
func (p *Pubrec) Type() uint8 {
	return TypeOfPubrec
}

// String returns the name of mqtt operation.
func (p *Pubrec) String() string {
	return "pubrec"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pubrel) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	buf.Write(reserveForHeader)
	writeUint16(buf, p.MessageID)

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfPubrel, buf, p.Header))
}

// Type returns the MQTT message type.
func (p *Pubrel) Type() uint8 {
	return TypeOfPubrel
}

// String returns the name of mqtt operation.
func (p *Pubrel) String() string {
	return "pubrel"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pubcomp) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	buf.Write(reserveForHeader)
	writeUint16(buf, p.MessageID)

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfPubcomp, buf, nil))
}

// Type returns the MQTT message type.
func (p *Pubcomp) Type() uint8 {
	return TypeOfPubcomp
}

// String returns the name of mqtt operation.
func (p *Pubcomp) String() string {
	return "pubcomp"
}

// EncodeTo writes the encoded message to the underlying writer.
func (s *Subscribe) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	buf.Write(reserveForHeader)
	writeUint16(buf, s.MessageID)
	for _, t := range s.Subscriptions {
		writeString(buf, t.Topic)
		buf.WriteByte(byte(t.Qos))
	}

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfSubscribe, buf, s.Header))
}

// Type returns the MQTT message type.
func (s *Subscribe) Type() uint8 {
	return TypeOfSubscribe
}

// String returns the name of mqtt operation.
func (s *Subscribe) String() string {
	return "sub"
}

// EncodeTo writes the encoded message to the underlying writer.
func (s *Suback) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	buf.Write(reserveForHeader)
	writeUint16(buf, s.MessageID)
	for _, q := range s.Qos {
		buf.WriteByte(byte(q))
	}

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfSuback, buf, nil))
}

// Type returns the MQTT message type.
func (s *Suback) Type() uint8 {
	return TypeOfSuback
}

// String returns the name of mqtt operation.
func (s *Suback) String() string {
	return "suback"
}

// EncodeTo writes the encoded message to the underlying writer.
func (u *Unsubscribe) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	buf.Write(reserveForHeader)
	writeUint16(buf, u.MessageID)
	for _, toptup := range u.Topics {
		writeString(buf, toptup.Topic)
	}

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfUnsubscribe, buf, u.Header))
}

// Type returns the MQTT message type.
func (u *Unsubscribe) Type() uint8 {
	return TypeOfUnsubscribe
}

// String returns the name of mqtt operation.
func (u *Unsubscribe) String() string {
	return "unsub"
}

// EncodeTo writes the encoded message to the underlying writer.
func (u *Unsuback) EncodeTo(w io.Writer) (int, error) {
	buf := buffers.Get()
	defer buffers.Put(buf)

	buf.Write(reserveForHeader)
	writeUint16(buf, u.MessageID)

	// Write to the underlying buffer
	return w.Write(encodeParts(TypeOfUnsuback, buf, nil))
}

// Type returns the MQTT message type.
func (u *Unsuback) Type() uint8 {
	return TypeOfUnsuback
}

// String returns the name of mqtt operation.
func (u *Unsuback) String() string {
	return "unsuback"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pingreq) EncodeTo(w io.Writer) (int, error) {
	return w.Write([]byte{0xc0, 0x0})
}

// Type returns the MQTT message type.
func (p *Pingreq) Type() uint8 {
	return TypeOfPingreq
}

// String returns the name of mqtt operation.
func (p *Pingreq) String() string {
	return "pingreq"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pingresp) EncodeTo(w io.Writer) (int, error) {
	return w.Write([]byte{0xd0, 0x0})
}

// Type returns the MQTT message type.
func (p *Pingresp) Type() uint8 {
	return TypeOfPingresp
}

// String returns the name of mqtt operation.
func (p *Pingresp) String() string {
	return "pingresp"
}

// EncodeTo writes the encoded message to the underlying writer.
func (d *Disconnect) EncodeTo(w io.Writer) (int, error) {
	return w.Write([]byte{0xe0, 0x0})
}

// Type returns the MQTT message type.
func (d *Disconnect) Type() uint8 {
	return TypeOfDisconnect
}

// String returns the name of mqtt operation.
func (d *Disconnect) String() string {
	return "disconnect"
}

// decodeStaticHeader decodes the header
func decodeStaticHeader(rdr io.Reader) (hdr *StaticHeader, length uint32, messageType uint8, err error) {
	b := make([]byte, 1)
	if _, err = io.ReadFull(rdr, b); err != nil {
		return nil, 0, 0, err
	}

	firstByte := b[0]
	messageType = (firstByte & 0xf0) >> 4

	// Set the header depending on the message type
	switch messageType {
	case TypeOfPublish, TypeOfSubscribe, TypeOfUnsubscribe, TypeOfPubrel:
		DUP := firstByte&0x08 > 0
		QOS := firstByte & 0x06 >> 1
		retain := firstByte&0x01 > 0

		hdr = &StaticHeader{
			DUP:    DUP,
			QOS:    QOS,
			Retain: retain,
		}
	}

	b[0] = 0x0 //b[0] ^= b[0]
	multiplier := uint32(1)
	digit := byte(0x80)

	// Read the length
	for (digit & 0x80) != 0 {
		if _, err = io.ReadFull(rdr, b); err != nil {
			return nil, 0, 0, err
		}

		digit = b[0]
		length += uint32(digit&0x7f) * multiplier
		multiplier *= 128
	}

	return hdr, uint32(length), messageType, nil
}

func decodeConnect(data []byte, hdr *StaticHeader) Message {
	//TODO: Decide how to recover rom invalid packets (offsets don't equal actual reading?)
	bookmark := uint32(0)

	protoname := readString(data, &bookmark)
	ver := uint8(data[bookmark])
	bookmark++
	flags := data[bookmark]
	bookmark++
	keepalive := readUint16(data, &bookmark)
	cliID := readString(data, &bookmark)
	connect := &Connect{
		ProtoName:      protoname,
		Version:        ver,
		KeepAlive:      keepalive,
		ClientID:       cliID,
		UsernameFlag:   flags&(1<<7) > 0,
		PasswordFlag:   flags&(1<<6) > 0,
		WillRetainFlag: flags&(1<<5) > 0,
		WillQOS:        (flags & (1 << 4)) + (flags & (1 << 3)),
		WillFlag:       flags&(1<<2) > 0,
		CleanSeshFlag:  flags&(1<<1) > 0,
	}

	if connect.WillFlag {
		connect.WillTopic = readString(data, &bookmark)
		connect.WillMessage = readString(data, &bookmark)
	}

	if connect.UsernameFlag {
		connect.Username = readString(data, &bookmark)
	}

	if connect.PasswordFlag {
		connect.Password = readString(data, &bookmark)
	}
	return connect
}

func decodeConnack(data []byte, hdr *StaticHeader) Message {
	//first byte is weird in connack
	bookmark := uint32(1)
	retcode := data[bookmark]

	return &Connack{
		ReturnCode: retcode,
	}
}

func decodePublish(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	topic := readString(data, &bookmark)
	var msgID uint16
	if hdr.QOS > 0 {
		msgID = readUint16(data, &bookmark)
	}

	return &Publish{
		Topic:     topic,
		Header:    hdr,
		Payload:   data[bookmark:],
		MessageID: msgID,
	}
}

func decodePuback(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Puback{
		MessageID: msgID,
	}
}

func decodePubrec(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubrec{
		MessageID: msgID,
	}
}

func decodePubrel(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubrel{
		Header:    hdr,
		MessageID: msgID,
	}
}

func decodePubcomp(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubcomp{
		MessageID: msgID,
	}
}

func decodeSubscribe(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	var topics []TopicQOSTuple
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t TopicQOSTuple
		t.Topic = readString(data, &bookmark)
		qos := data[bookmark]
		bookmark++
		t.Qos = uint8(qos)
		topics = append(topics, t)
	}
	return &Subscribe{
		Header:        hdr,
		MessageID:     msgID,
		Subscriptions: topics,
	}
}

func decodeSuback(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	var qoses []uint8
	maxlen := uint32(len(data))
	//is this efficient
	for bookmark < maxlen {
		qos := data[bookmark]
		bookmark++
		qoses = append(qoses, qos)
	}
	return &Suback{
		MessageID: msgID,
		Qos:       qoses,
	}
}

func decodeUnsubscribe(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	var topics []TopicQOSTuple
	msgID := readUint16(data, &bookmark)
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t TopicQOSTuple
		//		qos := data[bookmark]
		//		bookmark++
		t.Topic = readString(data, &bookmark)
		//		t.qos = uint8(qos)
		topics = append(topics, t)
	}
	return &Unsubscribe{
		Header:    hdr,
		MessageID: msgID,
		Topics:    topics,
	}
}

func decodeUnsuback(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Unsuback{
		MessageID: msgID,
	}
}

func decodePingreq(data []byte, hdr *StaticHeader) Message {
	return &Pingreq{}
}

func decodePingresp(data []byte, hdr *StaticHeader) Message {
	return &Pingresp{}
}

func decodeDisconnect(data []byte, hdr *StaticHeader) Message {
	return &Disconnect{}
}

// -------------------------------------------------------------
func writeString(buf *bytes.Buffer, s []byte) {
	strlen := uint16(len(s))
	writeUint16(buf, strlen)
	buf.Write(s)
}

func writeUint16(buf *bytes.Buffer, tupac uint16) {
	buf.WriteByte(byte((tupac & 0xff00) >> 8))
	buf.WriteByte(byte(tupac & 0x00ff))
}

func readString(b []byte, startsAt *uint32) []byte {
	l := readUint16(b, startsAt)
	v := b[*startsAt : uint32(l)+*startsAt]
	*startsAt += uint32(l)
	return v
}

func readUint16(b []byte, startsAt *uint32) uint16 {
	fst := uint16(b[*startsAt])
	*startsAt++
	snd := uint16(b[*startsAt])
	*startsAt++
	return (fst << 8) + snd
}

func boolToUInt8(v bool) uint8 {
	if v {
		return 0x1
	}

	return 0x0
}

// encodeLength encodes the length formatting (see: http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html#fixed-header)
// and tells us how many bytes it takes up.
func encodeLength(bodyLength uint32) (uint8, uint32) {
	if bodyLength == 0 {
		return 1, 0
	}

	var bitField uint32
	var numBytes uint8
	for bodyLength > 0 {
		bitField <<= 8
		dig := uint8(bodyLength % 128)
		bodyLength /= 128
		if bodyLength > 0 {
			dig = dig | 0x80
		}

		bitField |= uint32(dig)
		numBytes++
	}
	return numBytes, bitField
}
