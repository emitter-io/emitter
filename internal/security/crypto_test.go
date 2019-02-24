package security

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
			key:     "-ZgVnx1gr7BxxRCDsrEXBvmLVz86vGzs",
			channel: "cluster/#/",
			acl:     "rwsl",
		},
		{
			key:     "EbUlduEbUssgWueAWjkEZwdYG5YC0dGh",
			channel: "a/b/c/",
			acl:     "rwslpex",
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

		assert.Equal(t, uint32(989603869), key.Contract())
		encrypted, err := cipher.EncryptKey(key)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, tc.key, encrypted)
	}
}

func Test_encrypt(t *testing.T) {
	// Just an error test since everything is already covered by other tests
	cipher := &Cipher{key: [4]uint32{3443472288, 896798054, 972856492, 1831128908}}
	err := cipher.encrypt(make([]byte, 10))
	assert.Error(t, err)
}

func Test_decodeKey(t *testing.T) {
	// Just an error test since everything is already covered by other tests
	defer func(m [256]byte) { decodeMap = m }(decodeMap)
	for i := 0; i < len(decodeMap); i++ {
		decodeMap[i] = 1
	}

	in1 := []byte("#")
	_, err1 := decodeKey(make([]byte, 32), in1)
	assert.Error(t, err1)
	assert.NotEmpty(t, err1.Error())

	for i := 0; i < len(decodeMap); i++ {
		decodeMap[i] = byte(i)
	}

	for i := 0; i < 255; i++ {
		assert.NotPanics(t, func() {
			in2 := []byte{0, byte(i)}
			decodeKey(make([]byte, 32), in2)
		})
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
	tests := []struct {
		key      string
		expected [4]uint32
		err      bool
	}{
		{key: "zT83oDV0DWY5_JysbSTPTA", expected: [4]uint32{3443472288, 896798054, 972856492, 1831128908}},
		{key: "#%#%^", err: true},
		{key: "aaa", err: true},
	}

	for _, tc := range tests {
		cipher, err := NewCipher(tc.key)
		assert.Equal(t, tc.err, err != nil)
		if !tc.err {
			assert.EqualValues(t, tc.expected, cipher.key)
		}
	}
}

func TestGenerateKey(t *testing.T) {
	license, _ := ParseLicense("pLcaYvemMQOZR9o9sa5COWztxfAAAAAAAAAAAAAAAAI")
	cipher, _ := license.Cipher()
	masterKey, _ := cipher.DecryptKey([]byte("xEbaDPaICEwVhgdnl2rg_1DWi_MAg_3B"))

	tests := []struct {
		channel  string
		channels []string
		expected string
		err      bool
	}{
		{channel: "article1", err: true},
		{channel: "article1/", expected: "jhdrak0aHbbK6TbmyA391ndW3JwwgtNw", channels: []string{"article1/"}},
		{channel: "article1/#/", expected: "jhdrak0aHbbRDL1NpzzN4HdW3JwwgtNw", channels: []string{"article1/", "article1/a/", "article1/a/b/c/", "article1/+/a/b/c/"}},
	}

	for _, tc := range tests {
		key, err := cipher.GenerateKey(masterKey, tc.channel, AllowRead, time.Unix(0, 0), 1)
		assert.Equal(t, tc.err, err != nil)
		if tc.err {
			continue
		}

		// Assert the key
		assert.Equal(t, tc.expected, key)

		// Attempt to parse the key
		for _, c := range tc.channels {
			channel := ParseChannel([]byte(tc.expected + "/" + c))
			k, err := cipher.DecryptKey([]byte(key))
			assert.NoError(t, err)
			assert.True(t, k.ValidateChannel(channel))
		}

	}
}
