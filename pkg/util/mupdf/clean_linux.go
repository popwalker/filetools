package mupdf

/*
#cgo CFLAGS : -I${SRCDIR}/../../../thirdparty/mupdf/include
#cgo LDFLAGS: -L${SRCDIR}/../../../libs/mupdf/linux -lmupdf-pkcs7 -lmupdf -lmupdf-third -lmupdf-threads
#cgo LDFLAGS: -L/usr/lib/x86_64-linux-gnu -lm

#include "clean.h"
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
)

func pdfClean(infile, outfile string) error {
	in := C.CString(infile)
	defer C.free(unsafe.Pointer(in))

	out := C.CString(outfile)
	defer C.free(unsafe.Pointer(out))

	_ = C.pdfclean(in, out)
	//defer C.free(unsafe.Pointer(resCode))
	//fmt.Println(resCode)
	return nil
}
