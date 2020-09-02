package util

import (
	"reflect"
	"strings"
	"unsafe"
)

func UnsafeByte2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func UnsafeStr2Byte(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func isNeedTrim(r rune) bool {
	// 去除首位的换页符: \f,
	// Escape character: https://en.wikipedia.org/wiki/Escape_character
	if r == '\f' {
		return true
	}
	return false
}

func StringPurify(src string)string{
	cutset := []uint8{239, 187, 191}
	ret := strings.TrimLeft(src, string(cutset))
	ret = strings.TrimFunc(ret, isNeedTrim)
	ret = strings.NewReplacer("\n", "").Replace(src)
	return ret
}
