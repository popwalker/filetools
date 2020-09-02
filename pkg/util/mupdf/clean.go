package mupdf

import (
	"invtools/utils"

	"invtools/utils/errors"
)

/*
clean.go rewrite(repair) pdf file
*/

// PdfRepair pdf修复
func PdfRepair(infile, outfile string) error {
	if ok := utils.CheckFileIsExist(infile); !ok {
		return errors.Errorf(nil, "infile not exists")
	}

	return pdfClean(infile, outfile)
}
