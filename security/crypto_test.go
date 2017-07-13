package security

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

type x struct {
	test string
}

func BenchmarkDecryptKey(b *testing.B) {
	cipher := &Cipher{key: [4]uint32{3443472288, 896798054, 972856492, 1831128908}}
	key := "0TJnt4yZPL73zt35h1UTIFsYBLetyD_g"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cipher.DecryptKey([]byte(key))
	}
}

func TestEncryptDecrypt(t *testing.T) {

	// license: zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI
	// secret key: kBCZch5re3Ue-kpG1Aa8Vo7BYvXZ3UwR
	cipher := &Cipher{key: [4]uint32{3443472288, 896798054, 972856492, 1831128908}}
	tests := []struct {
		key     string
		channel string
		acl     string
		err     bool
	}{
		{
			key:     "0TJnt4yZPL73zt35h1UTIFsYBLetyD_g",
			channel: "emitter",
			acl:     "rsp",
		},
		{
			key:     "xm54Sj0srWlSEctra-yU6ZA6Z2e6pp7c",
			channel: "a/b/c",
			acl:     "rw",
		},
		{
			key:     "YXiszDIuAbkiyJfG-J1YpAwI0jUAbXW_",
			channel: "a/b/c",
			acl:     "sl",
		},
		{
			key: "",
			err: true,
		},
		{
			key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa*",
			err: true,
		},
	}

	for _, tc := range tests {
		key, err := cipher.DecryptKey([]byte(tc.key))
		assert.Equal(t, tc.err, err != nil, tc.key)
		if tc.err {
			continue
		}

		assert.Equal(t, int32(989603869), key.Contract())
		encrypted, err := cipher.EncryptKey(key)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, tc.key, encrypted)
	}
}

func BenchmarkBase64(b *testing.B) {
	v := []byte("0TJnt4yZPL73zt35h1UTIFsYBLetyD_g")
	o := make([]byte, 24)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decodeKey(o, v)
	}
}

func TestBase64Decode(t *testing.T) {
	v := []byte("0TJnt4yZPL73zt35h1UTIFsYBLetyD_g")
	o1 := make([]byte, 24)
	o2 := make([]byte, 24)

	n1, e1 := decodeKey(o1, v)
	n2, e2 := base64.RawURLEncoding.Decode(o2, v)

	assert.Equal(t, o1, o2)
	assert.Equal(t, n1, n2)
	assert.Equal(t, e1, e2)

}

func TestBase64SelfDecode(t *testing.T) {
	v := []byte("0TJnt4yZPL73zt35h1UTIFsYBLetyD_g")
	o1 := []byte("0TJnt4yZPL73zt35h1UTIFsYBLetyD_g")
	o2 := make([]byte, 24)

	n1, e1 := decodeKey(o1, o1)
	n2, e2 := base64.RawURLEncoding.Decode(o2, v)

	assert.Equal(t, o1[:24], o2)
	assert.Equal(t, n1, n2)
	assert.Equal(t, e1, e2)
}

func TestNewCipher(t *testing.T) {
	license, err := ParseLicense("zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI")
	if err != nil {
		t.Error(err)
	}

	cipher, err := NewCipher(license.EncryptionKey)
	if err != nil {
		t.Error(err)
	}

	expected := [4]uint32{3443472288, 896798054, 972856492, 1831128908}

	assert.EqualValues(t, expected, cipher.key)

}
