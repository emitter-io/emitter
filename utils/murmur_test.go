package utils

import (
	"testing"
)

//  16.1 ns/op             0 B/op          0 allocs/op
func BenchmarkGetHash(b *testing.B) {
	v := []byte("a/b/c/d/e/f/g/h/this/is/emitter")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetHash(v)
	}
}

func TestGetHash(t *testing.T) {
	h := GetHash([]byte("+"))
	if h != 1815237614 {
		t.Errorf("Hash %d is not equal to %d", h, 1815237614)
	}
}
