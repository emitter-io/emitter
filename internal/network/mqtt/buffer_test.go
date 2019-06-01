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

package mqtt

import (
	"bytes"
	"io"
	"testing"
)

// BenchmarkPublishEncode-8   	20000000	        93.6 ns/op	       0 B/op	       0 allocs/op
func BenchmarkPublishEncode(b *testing.B) {
	benchmarkPacketEncode(b, &Publish{
		Header: Header{
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

// BenchmarkPublishDecode-8   	 5000000	       243 ns/op	     960 B/op	       2 allocs/op
func BenchmarkPublishDecode(b *testing.B) {
	benchmarkPacketDecode(b, &Publish{
		Header: Header{
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
		_, err := DecodePacket(reader, 65536)
		if err != nil {
			b.Error(err)
		}
	}
}
