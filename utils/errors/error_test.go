package errors

import (
	"testing"
)

func TestErrorf(t *testing.T) {
	err := Errorf(nil, "this is a error message")
	t.Log(err)

	f1 := func() {
		err = Errorf(nil, "error message in closure")
	}
	f1()
	t.Log(err)

	err = Errorf(err, "wrap error")
	t.Log(err)
}

func TestExample(t *testing.T) {
	func1()
}

func TestExample2(t *testing.T) {
	func11()
}

// BenchmarkErrorfViaFmtFprintf-8            200000              8236 ns/op            3609 B/op         24 allocs/op
// BenchmarkErrorfViaWriteString-8           200000              7240 ns/op            3457 B/op         13 allocs/op
// 去掉regexp后
// BenchmarkErrorfViaWriteString-8           500000              3272 ns/op            3073 B/op          8 allocs/op
func BenchmarkErrorfViaWriteString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func2().(*Err).Error()
	}
}
