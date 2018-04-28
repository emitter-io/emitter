package address

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	assert.Equal(t, uint64(0x313233343536), encode([]byte("123456")))
}

func TestGlobals(t *testing.T) {
	assert.NotEqual(t, "", External().String())
	assert.NotEqual(t, Fingerprint(0), Hardware())
	assert.NotEqual(t, "", Hardware().String())
	assert.NotEqual(t, "", Hardware().Hex())
}

func TestParse(t *testing.T) {
	tests := []struct {
		addr string
		err  bool
	}{
		{addr: "private"},
		{addr: "private:8080"},
		{addr: "public"},
		{addr: "10.0.0.1:4000"},
		{addr: "", err: true},
		{addr: ":8000"},
	}

	for _, tc := range tests {
		t.Run("Parsing"+tc.addr, func(*testing.T) {
			_, err := Parse(tc.addr, 80)
			assert.Equal(t, tc.err, err != nil)
			if err != nil {
				println(err.Error())
			}
		})
	}
}
