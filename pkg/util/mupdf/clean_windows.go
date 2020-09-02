package mupdf

/*
#cgo CFLAGS : -I${SRCDIR}/../../../thirdparty/mupdf/include
#cgo LDFLAGS: -L${SRCDIR}/../../../libs/mupdf/windows -lmupdf -lmupdfthird

#include "clean.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func pdfClean(infile, outfile string) error {
	in := C.CString(infile)
	defer C.free(unsafe.Pointer(in))

	out := C.CString(outfile)
	defer C.free(unsafe.Pointer(out))

	resCode := C.pdfclean(in, out)
	//defer C.free(unsafe.Pointer(resCode))
	fmt.Println(resCode)
	return nil
}
