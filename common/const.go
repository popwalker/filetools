package common

import (
	"strings"
)

const (
	ExtPDF   = ".pdf"
	ExtCsv   = ".csv"
	ExtExecl = ".xlsx"
	ExtTxt   = ".txt"
	ExtZip   = ".zip"
	ExtPng   = ".png"
	ExtJpeg  = ".jpeg"
	ExtJpg   = ".jpg"
)

const (
	CodeTypeQRCode     = "qrcode"
	CodeTypeBarcode128 = "barcode128"
)

var AllowedCsvExts = []string{
	ExtCsv, ExtExecl,
}

func AllowedOutputTable(e string) bool {
	for _, ext := range AllowedCsvExts {
		if strings.EqualFold(e, ext) {
			return true
		}
	}
	return false
}

var (
	// RunningDetective Detective which running current now
	RunningDetective string

	CurrentDir string

	HomeDir string
)
