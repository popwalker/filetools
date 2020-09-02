package util

import (
	"bytes"
	"strings"
	"testing"
)

var (
	ts = strings.Repeat("a", 1024)
	tb = bytes.Repeat([]byte("a"), 1024)
)

func BenchmarkUnsafeByte2str(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UnsafeByte2str(tb)
	}
}
func BenchmarkByte2str(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = string(tb)
	}
}

func BenchmarkUnsafeStr2Byte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UnsafeStr2Byte(ts)
	}
}
func BenchmarkStr2Byte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = []byte(ts)
	}
}
