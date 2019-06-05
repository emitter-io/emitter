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

package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkParseChannelWithOptions(b *testing.B) {
	in := []byte("xm54Sj0srWlSEctra-yU6ZA6Z2e6pp7c/a/roman/is/da/best/?opt1=true&opt2=false")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseChannel(in)
	}
}

func BenchmarkParseChannelStatic(b *testing.B) {
	in := []byte("xm54Sj0srWlSEctra-yU6ZA6Z2e6pp7c/a/roman/is/da/best/")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseChannel(in)
	}
}

func TestParseChannel(t *testing.T) {
	tests := []struct {
		k  string
		ch string
		o  []string
		t  uint8
	}{
		{k: "emitter", ch: "a/", t: ChannelStatic},
		{k: "emitter", ch: "a/b/c/", t: ChannelStatic},
		{k: "emitter", ch: "test-channel/", t: ChannelStatic},
		{k: "emitter", ch: "test-channel/+/and-more/", t: ChannelWildcard},
		{k: "emitter", ch: "a/-/x/", t: ChannelStatic},
		{k: "emitter", ch: "a/b/c/d/", t: ChannelStatic},
		{k: "emitter", ch: "a/b/c/+/", t: ChannelWildcard},
		{k: "emitter", ch: "a/+/c/+/", t: ChannelWildcard},
		{k: "emitter", ch: "b/+/", t: ChannelWildcard},
		{k: "0TJnt4yZPL73zt35h1UTIFsYBLetyD_g", ch: "emitter/", o: []string{"test=true", "something=7"}, t: ChannelStatic},
		{k: "emitter", ch: "a/b/c/d/", o: []string{"test=true", "something=7"}, t: ChannelStatic},
		{k: "emitter", ch: "a/b/c/d/", o: []string{"req=13", "something=7"}, t: ChannelStatic},

		// Invalid channels
		{t: ChannelInvalid},
		{k: "emitter", ch: "a/@/x/", t: ChannelInvalid},
		{k: "emitter", ch: "a", t: ChannelInvalid},
		{k: "emitter", ch: "a/b/c", t: ChannelInvalid},
		{k: "emitter", ch: "a//b/", t: ChannelInvalid},
		{k: "emitter", ch: "a//////b/c", t: ChannelInvalid},
		{k: "emitter", ch: "*", t: ChannelInvalid},
		{k: "emitter", ch: "+", t: ChannelInvalid},
		{k: "emitter", ch: "a/+", t: ChannelInvalid},
		{k: "emitter", ch: "b/+", t: ChannelInvalid},
		{k: "emitter", ch: "b/*+/", t: ChannelInvalid},
		{k: "emitter", ch: "b/+a/", t: ChannelInvalid},
		{k: "emitter", ch: "", t: ChannelInvalid},
		{k: "emitter", ch: "/", t: ChannelInvalid},
		{k: "emitter", ch: "//", t: ChannelInvalid},
		{k: "emitter", ch: "a//", t: ChannelInvalid},
		{k: "emitter", ch: "a/b/c/d/", o: []string{"test=true", "something=7", "more=_"}, t: ChannelInvalid},
		{k: "emitter", ch: "a/b/c/d/", o: []string{"test==true"}, t: ChannelInvalid},
		{k: "emitter", ch: "a/b/c/d/", o: []string{"te_st==true"}, t: ChannelInvalid},
		{k: "emitter", ch: "a/", o: []string{"=true"}, t: ChannelInvalid},
		{k: "emitter", ch: "a/", o: []string{"test="}, t: ChannelInvalid},
		//		{k: "emitter", ch: "a/b/c/d", o: []string{"test=="}, err: true},
	}

	for _, tc := range tests {
		// First we need to build the input to parse
		in := tc.k + "/" + tc.ch
		if len(tc.o) > 0 {
			in += "?"
			in += strings.Join(tc.o, "&")
		}

		// Parse the channel now
		out := ParseChannel([]byte(in))
		assert.Equal(t, tc.t, out.ChannelType, "input: "+in)
		if tc.t != ChannelInvalid && out.ChannelType != ChannelInvalid {

			// Make sure this always ends with a trailing slash
			if !strings.HasSuffix(tc.ch, "/") {
				tc.ch += "/"
			}

			//assert.Equal(t, ChannelStatic, out.Type)
			assert.Equal(t, tc.k, string(out.Key), "input: "+in)
			assert.Equal(t, tc.ch, string(out.Channel), "input: "+in)

			// Check the options
			for _, opt := range tc.o {
				target := strings.Split(opt, "=")[0]

				found := false
				for _, kvp := range out.Options {
					if kvp.Key == target {
						found = true
						assert.Equal(t, strings.Split(opt, "=")[1], kvp.Value)
					}
				}

				assert.Equal(t, true, found, "unable to find key = "+target)
			}
		}
	}
}

func TestGetChannelExclude(t *testing.T) {
	tests := []struct {
		channel string
		ok      bool
	}{
		{channel: "emitter/a/?me=0", ok: true},
		{channel: "emitter/a/?me=12000000", ok: false},
		{channel: "emitter/a/?me=1200a", ok: false},
		{channel: "emitter/a/?me=-1", ok: false},
		{channel: "emitter/a/", ok: false},
	}

	for _, tc := range tests {
		channel := ParseChannel([]byte(tc.channel))
		excludeMe := channel.Exclude()

		assert.Equal(t, excludeMe, tc.ok)
	}
}

func TestGetChannelTTL(t *testing.T) {
	tests := []struct {
		channel string
		ttl     int64
		ok      bool
	}{
		{channel: "emitter/a/?ttl=42&abc=9", ttl: 42, ok: true},
		{channel: "emitter/a/?ttl=1200", ttl: 1200, ok: true},
		{channel: "emitter/a/?ttl=1200a", ok: false},
		{channel: "emitter/a/", ok: false},
	}

	for _, tc := range tests {
		channel := ParseChannel([]byte(tc.channel))
		ttl, hasValue := channel.TTL()

		assert.Equal(t, tc.ttl, ttl)
		assert.Equal(t, hasValue, tc.ok)
	}
}

func TestGetChannelLast(t *testing.T) {
	tests := []struct {
		channel string
		last    int64
		ok      bool
	}{
		{channel: "emitter/a/?last=42&abc=9", last: 42, ok: true},
		{channel: "emitter/a/?last=1200", last: 1200, ok: true},
		{channel: "emitter/a/?last=1200a", ok: false},
		{channel: "emitter/a/", ok: false},
	}

	for _, tc := range tests {
		channel := ParseChannel([]byte(tc.channel))
		last, hasValue := channel.Last()

		assert.Equal(t, tc.last, last)
		assert.Equal(t, hasValue, tc.ok)
	}
}

func TestGetChannelWindow(t *testing.T) {
	tests := []struct {
		channel string
		t0      int64
		t1      int64
	}{
		{channel: "emitter/a/?from=42&abc=9", t0: 0, t1: 0},
		{channel: "emitter/a/?from=1200", t0: 0, t1: 0},
		{channel: "emitter/a/?from=1200&until=2550", t0: 0, t1: 0},
		{channel: "emitter/a/?from=1200a", t0: 0, t1: 0},
		{channel: "emitter/a/", t0: 0, t1: 0},
		{channel: "emitter/a/?from=1514764800&until=1514764900", t0: 1514764800, t1: 1514764900},
		{channel: "emitter/a/?from=1514764800", t0: 1514764800, t1: 0},
		{channel: "emitter/a/?until=1514764900", t0: 0, t1: 1514764900},
		{channel: "emitter/a/?from=1514764800&until=3029529610", t0: 1514764800, t1: 0},
		{channel: "emitter/a/?from=1514764900&until=1514764800", t0: 1514764900, t1: 1514764800},
	}

	for _, tc := range tests {
		channel := ParseChannel([]byte(tc.channel))
		t0, t1 := channel.Window()

		assert.Equal(t, tc.t0, t0.Unix())
		assert.Equal(t, tc.t1, t1.Unix())
	}
}

func TestGetChannelTarget(t *testing.T) {
	tests := []struct {
		channel string
		target  uint32
	}{
		{channel: "emitter/a/?ttl=42&abc=9", target: 0xc103eab3},
		{channel: "emitter/$share/a/b/c/", target: 1480642916},
	}

	for _, tc := range tests {
		channel := ParseChannel([]byte(tc.channel))
		target := channel.Target()

		assert.Equal(t, tc.target, target)
	}
}

func TestMakeChannel(t *testing.T) {
	tests := []struct {
		key     string
		channel string
	}{
		{key: "key1", channel: "emitter/a/"},
	}

	for _, tc := range tests {
		channel := MakeChannel(tc.key, tc.channel)
		assert.Equal(t, tc.key, string(channel.Key))
		assert.Equal(t, tc.channel, string(channel.Channel))
	}
}

func TestChannelString(t *testing.T) {
	tests := []struct {
		channel string
	}{
		{channel: "emitter/a/?last=42&abc=9"},
		{channel: "emitter/a/?last=1200"},
		{channel: "emitter/a/?last=1200a"},
		{channel: "emitter/a/"},
	}

	for _, tc := range tests {
		channel := ParseChannel([]byte(tc.channel))
		assert.Equal(t, tc.channel, channel.String())
	}
}
