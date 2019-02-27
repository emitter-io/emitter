package mqtt

import (
	"bytes"
	"io"
	"log"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type devNull struct{}

func (dn devNull) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func BenchmarkPublishEncode(b *testing.B) {
	benchmarkPacketEncode(b, &Publish{
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
		Payload:   []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum mattis enim nec lacinia pharetra. Fusce a nibh augue. Donec lectus felis, feugiat id pellentesque semper, tincidunt in mi. Nunc molestie facilisis magna, eget imperdiet enim pulvinar sed. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus hendrerit nibh at vestibulum imperdiet. Vivamus ac blandit augue, at fermentum felis. Praesent elementum eu nisl vel egestas. Cras vestibulum suscipit pulvinar. Praesent vel quam id risus dictum suscipit. Nunc porta massa eget rhoncus varius. Nam sed lorem orci. Quisque odio mi, pretium in eros id, convallis luctus purus. Curabitur in placerat dolor. Duis sit amet tellus molestie, auctor nibh placerat, condimentum nunc. Etiam placerat leo dapibus cursus laoreet."),
		Topic:     []byte("a/b/c"),
		MessageID: 1,
	})
}

func benchmarkPacketEncode(b *testing.B, packet Message) {
	w := devNull{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = packet.EncodeTo(w)
	}
}

func BenchmarkPublishDecode(b *testing.B) {
	benchmarkPacketDecode(b, &Publish{
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
		Payload:   []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum mattis enim nec lacinia pharetra. Fusce a nibh augue. Donec lectus felis, feugiat id pellentesque semper, tincidunt in mi. Nunc molestie facilisis magna, eget imperdiet enim pulvinar sed. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus hendrerit nibh at vestibulum imperdiet. Vivamus ac blandit augue, at fermentum felis. Praesent elementum eu nisl vel egestas. Cras vestibulum suscipit pulvinar. Praesent vel quam id risus dictum suscipit. Nunc porta massa eget rhoncus varius. Nam sed lorem orci. Quisque odio mi, pretium in eros id, convallis luctus purus. Curabitur in placerat dolor. Duis sit amet tellus molestie, auctor nibh placerat, condimentum nunc. Etiam placerat leo dapibus cursus laoreet."),
		Topic:     []byte("a/b/c"),
		MessageID: 1,
	})
}

func BenchmarkConnectDecode(b *testing.B) {
	benchmarkPacketDecode(b, &Connect{
		ProtoName:      []byte("MQTsdp"),
		Version:        3,
		UsernameFlag:   false,
		PasswordFlag:   false,
		WillRetainFlag: false,
		WillQOS:        0,
		WillFlag:       false,
		CleanSeshFlag:  true,
		KeepAlive:      30,
		ClientID:       []byte("13241"),
		WillTopic:      []byte("a/b/c"),
		WillMessage:    []byte("tommy this and tommy that and tommy ow's yer soul"),
		Username:       []byte("Username"),
		Password:       []byte("Password"),
	})
}

func benchmarkPacketDecode(b *testing.B, packet Message) {
	slc := bytes.NewBuffer([]byte{})
	_, _ = packet.EncodeTo(slc)
	reader := bytes.NewReader(slc.Bytes())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Seek(0, io.SeekStart)
		_, err := DecodePacket(reader,65536)
		if err != nil {
			b.Error(err)
		}
	}
}

func Test_LargePacket(t *testing.T) {
	pay := make([]byte, 65536-10)
	for i := range pay {
		pay[i] = 0x0f
	}

	pub := &Publish{
		Header: &StaticHeader{
			QOS:    0,
			Retain: false,
			DUP:    false,
		},
		Payload:   pay,
		Topic:     []byte("a/b/c"),
		MessageID: 69,
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			slc := bytes.NewBuffer([]byte{})
			_, _ = pub.EncodeTo(slc)
			_, err := DecodePacket(slc,65536)
			if err != nil {
				t.Error(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func encodeTestHelper(toEncode Message) bool {
	buf := bytes.NewBuffer([]byte{})
	_, _ = toEncode.EncodeTo(buf)
	msg, err := DecodePacket(buf,65536)
	if err != nil {
		log.Printf("error in here %+v\n", err.Error())
		return false
	}
	match := false
	switch msg.(type) {
	case *Connect:
		match = msg.Type() == TypeOfConnect
	case *Connack:
		match = msg.Type() == TypeOfConnack
	case *Publish:
		match = msg.Type() == TypeOfPublish
	case *Pubrec:
		match = msg.Type() == TypeOfPubrec
	case *Puback:
		match = msg.Type() == TypeOfPuback
	case *Pubrel:
		match = msg.Type() == TypeOfPubrel
	case *Pubcomp:
		match = msg.Type() == TypeOfPubcomp
	case *Subscribe:
		match = msg.Type() == TypeOfSubscribe
	case *Suback:
		match = msg.Type() == TypeOfSuback
	case *Unsubscribe:
		match = msg.Type() == TypeOfUnsubscribe
	case *Unsuback:
		match = msg.Type() == TypeOfUnsuback
	case *Pingreq:
		match = msg.Type() == TypeOfPingreq
	case *Pingresp:
		match = msg.Type() == TypeOfPingresp
	case *Disconnect:
		match = msg.Type() == TypeOfDisconnect
	}
	if match != true {
		return false
	}
	return reflect.DeepEqual(toEncode, msg)
}

func Test_Connect(t *testing.T) {
	testPkt := &Connect{
		ProtoName:      []byte("MQTsdp"),
		Version:        3,
		UsernameFlag:   true,
		PasswordFlag:   true,
		WillRetainFlag: true,
		WillQOS:        0,
		WillFlag:       true,
		CleanSeshFlag:  true,
		KeepAlive:      30,
		ClientID:       []byte("420"),
		WillTopic:      []byte("a/b/c"),
		WillMessage:    []byte("tommy this and tommy that and tommy ow's yer soul"),
		Username:       []byte("Username"),
		Password:       []byte("Password is my password"),
	}

	assert.Equal(t, "connect", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode connect failed")
	}
}

func Test_Connack(t *testing.T) {
	testPkt := &Connack{
		ReturnCode: 0x04,
	}

	assert.Equal(t, "connack", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode connack failed")
	}
}

func Test_Publish(t *testing.T) {
	testPkt := &Publish{
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
		Payload:   []byte("tommy this and tommy that"),
		Topic:     []byte("a/b/c"),
		MessageID: 69,
	}

	assert.Equal(t, "pub", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode publish failed")
	}
}

func Test_Publish2(t *testing.T) {
	testPkt := &Publish{
		Header: &StaticHeader{
			QOS:    2,
			Retain: false,
			DUP:    false,
		},
		Payload:   []byte("A thin red line of 'eroes"),
		Topic:     []byte("a/b/c"),
		MessageID: 69,
	}
	buf := bytes.NewBuffer([]byte{})
	_, _ = testPkt.EncodeTo(buf)
	msg, err := DecodePacket(buf,65536)
	if err != nil {
		t.Error(err.Error())
	}
	if msg.(*Publish).Header.QOS != testPkt.Header.QOS {
		t.Error("Encode/decode failed on test publish 2")
	}
}

func Test_Publish_WithUnicodeDecoding(t *testing.T) {
	pay := []byte("hello earth 😁, good evening")
	testPkt := &Publish{
		Header: &StaticHeader{
			QOS:    2,
			Retain: false,
			DUP:    false,
		},
		Payload:   pay,
		Topic:     []byte("a/b/c"),
		MessageID: 69,
	}

	buf := bytes.NewBuffer([]byte{})
	_, _ = testPkt.EncodeTo(buf)

	msg, err := DecodePacket(buf,65536)
	if err != nil {
		t.Error(err.Error())
	}
	if msg.(*Publish).Header.QOS != testPkt.Header.QOS {
		t.Error("Encode/decode failed on test publish 2")
	}

	if !bytes.Equal(msg.(*Publish).Payload, pay) {
		log.Println(pay)
		log.Println(msg.(*Publish).Payload)

		log.Println(string(pay))
		log.Println(string(msg.(*Publish).Payload))

		t.Error("Invalid encoding")
	}
}

func Test_Puback(t *testing.T) {
	testPkt := &Puback{
		MessageID: 0xbeef,
	}

	assert.Equal(t, "puback", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode puback failed")
	}
}

func Test_Pubrec(t *testing.T) {
	testPkt := &Pubrec{
		MessageID: 0xbeef,
	}

	assert.Equal(t, "pubrec", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode pubrec failed")
	}
}

func Test_Pubrel(t *testing.T) {
	testPkt := &Pubrel{
		MessageID: 0xbeef,
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
	}

	assert.Equal(t, "pubrel", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode pubrel failed")
	}
}

func Test_Pubcomp(t *testing.T) {
	testPkt := &Pubcomp{
		MessageID: 0xbeef,
	}

	assert.Equal(t, "pubcomp", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode pubcomp failed")
	}
}

func Test_Subscribe(t *testing.T) {
	testPkt := &Subscribe{
		MessageID: 0xbeef,
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
		Subscriptions: []TopicQOSTuple{
			{
				Qos:   0,
				Topic: []byte("a/b/c"),
			},
		},
	}

	assert.Equal(t, "sub", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode subscribe failed")
	}
}

func Test_Suback(t *testing.T) {
	testPkt := &Suback{
		MessageID: 0xbeef,
		Qos:       []uint8{0, 0, 1},
	}

	assert.Equal(t, "suback", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode suback failed")
	}
}

func Test_UnSubscribe(t *testing.T) {
	testPkt := &Unsubscribe{
		MessageID: 0xbeef,
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
		Topics: []TopicQOSTuple{
			{
				Qos:   0,
				Topic: []byte("a/b/c"),
			},
		},
	}

	assert.Equal(t, "unsub", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode unsubscribe failed")
	}
}

func Test_Unsuback(t *testing.T) {
	testPkt := &Unsuback{
		MessageID: 0xbeef,
	}

	assert.Equal(t, "unsuback", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode unsuback failed")
	}
}

func Test_PingReq(t *testing.T) {
	testPkt := &Pingreq{}
	assert.Equal(t, "pingreq", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode pingreq failed")
	}
}

func Test_PingResp(t *testing.T) {
	testPkt := &Pingresp{}
	assert.Equal(t, "pingresp", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode pingresp failed")
	}
}

func Test_Disconnect(t *testing.T) {
	testPkt := &Disconnect{}
	assert.Equal(t, "disconnect", testPkt.String())
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode disconnect failed")
	}
}

func Test_encodeLength(t *testing.T) {
	test := func(testval, expecField uint32, expecLeng uint8, t *testing.T) {
		fmtStr := "invalid response from encodeLength field %b leng %d, expected field %b expected value %d\n"
		leng, field := encodeLength(testval)
		if field != expecField || leng != expecLeng {
			t.Errorf(fmtStr, field, leng, expecField, expecLeng)
		}
	}

	test(0, 0x0, 1, t)
	test(1, 0x1, 1, t)
	test(127, 0x7f, 1, t)
	test(128, 0x8001, 2, t)
	test(16383, 0xff7f, 2, t)
	test(16384, 0x808001, 3, t)
	test(2097151, 0xffff7f, 3, t)
	test(2097152, 0x80808001, 4, t)
	test(268435455, 0xffffff7f, 4, t)
}

func Test_DecodeLength(t *testing.T) {
	tst := func(tst uint32, t *testing.T) {
		_, enclen := encodeLength(tst)
		var byteee [4]byte
		byteee[0] = byte(enclen >> 24)
		byteee[1] = byte(enclen >> 16)
		byteee[2] = byte(enclen >> 8)
		byteee[3] = byte(enclen)
		if res := decodeLen(byteee[:]); res != tst {
			t.Errorf("expected %d and got %d\n", tst, res)
		}
	}

	test := func(expecField, testval uint32, expecLeng uint8, t *testing.T) {
		fmtStr := "invalid response from encodeLength field %b leng %d, expected field %b expected value %d\n"
		var blah [4]byte
		blah[0] = byte(testval >> 24)
		blah[1] = byte(testval >> 16)
		blah[2] = byte(testval >> 8)
		blah[3] = byte(testval)

		field := decodeLen(blah[:])
		if field != expecField {
			t.Errorf(fmtStr, field, 0, expecField, expecLeng)
		}
	}
	tst(986889, t)
	tst(0, t)
	tst(1, t)
	tst(127, t)
	tst(128, t)
	tst(16383, t)
	tst(16384, t)
	tst(209715, t)
	tst(2097152, t)
	tst(268435455, t)
	test(0, 0x0, 1, t)
	test(1, 0x1, 1, t)
	test(127, 0x7f, 1, t)
	test(128, 0x8001, 2, t)
	test(16383, 0xff7f, 2, t)
	test(16384, 0x808001, 3, t)
	test(2097151, 0xffff7f, 3, t)
	test(2097152, 0x80808001, 4, t)
	test(268435455, 0xffffff7f, 4, t)

}

func decodeLen(field []byte) uint32 {
	//sadly I have to ape decoding length
	multiplier := uint32(1)
	length := uint32(0) //signed for great justice?
	digit := byte(0x80)
	rdr := bytes.NewBuffer(field)

	//since we're writing the byte pattern to a 4 byte slice, no matter what the actual size, we have to skip the leftmost empty bytes
	var b [1]byte
	steps := 1
	for (digit & 0x80) != 0 {
		_, err := io.ReadFull(rdr, b[:])
		if err != nil {
			log.Println(digit, steps)
			panic(err.Error())
		}
		if b[0] == 0 {
			if steps == 4 {
				return 0
			}

			steps++
			continue
		}
		steps++
		digit = b[0]

		length += uint32(digit&0x7f) * multiplier
		multiplier *= 128

	}
	return length
}
