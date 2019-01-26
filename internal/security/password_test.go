package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePassword(t *testing.T) {
	tests := []struct {
		input   string
		expect  string
		success bool
	}{
		{input: "dial://emitter/a/?ttl=42&abc=9", expect: "a/", success: true},
		{input: "dial://emitter/a/?ttl=1200", expect: "a/", success: true},
		{input: "dial://emitter/a/?ttl=1200a", expect: "a/", success: true},
		{input: "dial://emitter/a/b/c/", expect: "a/b/c/", success: true},

		{input: "auto:/test"},
		{input: "auto://"},
		{input: "auto://--------2345205982)@#(%&*@#//)//%"},
		{input: "dial://emitter/+/"},
		{input: "emitter/a/?ttl=42&abc=9"},
		{input: "emitter/a/?ttl=1200"},
		{input: "emitter/a/?ttl=1200a"},
		{input: "emitter/a/"},
		{input: "err://emitter/a/?ttl=42&abc=9"},
		{input: "err://emitter/a/?ttl=1200"},
		{input: "err://emitter/a/?ttl=1200a"},
		{input: "err://emitter/a/"},
	}

	for _, tc := range tests {
		scheme, channel := ParsePassword(tc.input)
		assert.Equal(t, scheme != "", tc.success, tc.input)
		if tc.expect != "" {
			assert.Equal(t, tc.expect, string(channel.Channel), tc.input)
		}
	}
}
