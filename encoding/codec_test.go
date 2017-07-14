package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Ssid    []uint32
	Channel string
}

func TestEncodeDecode(t *testing.T) {
	v := testStruct{
		Ssid:    []uint32{1, 2, 3},
		Channel: "Hello World",
	}

	encoded, err := Encode(1, v)
	assert.NoError(t, err)

	o := testStruct{}
	err = Decode(encoded[1:], &o)
	assert.NoError(t, err)
	assert.Equal(t, v, o)

}
