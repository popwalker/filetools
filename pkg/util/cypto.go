package util

import (
	"crypto/md5"
	"fmt"
)

func ComputeMd5String(s string) string {
	data := UnsafeStr2Byte(s)
	return fmt.Sprintf("%x", md5.Sum(data))
}
