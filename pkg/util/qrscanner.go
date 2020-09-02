package util

import (
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"strings"

	"invtools/common"

	"invtools/utils/errors"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/makiuchi-d/gozxing/qrcode"
	multiqrcode "github.com/makiuchi-d/gozxing/multi/qrcode"
)

func ImgDecode(ext string, f *os.File) (image.Image, error) {
	if f == nil {
		return nil, errors.Errorf(nil, "param f is nil")
	}

	switch ext {
	case common.ExtPng:
		return png.Decode(f)
	case common.ExtJpeg, common.ExtJpg:
		return jpeg.Decode(f)
	default:
		return nil, errors.Errorf(nil, "unsupported image file ext")
	}
}

// QrCodeScan scan qrcode image
func QrCodeScan(input string) (code string, err error) {
	file, err := os.Open(input)
	if err != nil {
		err = errors.Errorf(err, "open file failed")
		return
	}

	img, err := ImgDecode(path.Ext(input), file)
	if err != nil {
		err = errors.Errorf(err, "decode image file failed")
		return
	}

	// prepare BinaryBitmap
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		err = errors.Errorf(err, "gozxing NewBinaryBitmapFromImage failed ")
		return
	}

	// decode image
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		if strings.Contains(err.Error(), "NotFoundException") || strings.Contains(err.Error(), "FormatException"){
			multiQrReader := multiqrcode.NewQRCodeMultiReader()
			result, er := multiQrReader.DecodeMultipleWithoutHint(bmp)
			if er != nil {
				err = errors.Errorf(er, "gozxing/multi/qrcode reader decode qrcode image failed")
				return
			}
			if len(result) > 0{
				return result[0].String(), nil
			}
			return "", errors.Errorf(nil, "gozxing/multi/qrcode got multi result")
		}
		err = errors.Errorf(err, "gozxing/qrcode reader decode qrcode image failed")
		return
	}

	return result.String(), nil
}

// Barcode128Scan scan barcode128 image
func Barcode128Scan(input string) (code string, err error) {
	file, err := os.Open(input)
	if err != nil {
		err = errors.Errorf(err, "open file failed")
		return
	}

	img, err := ImgDecode(path.Ext(input), file)
	if err != nil {
		err = errors.Errorf(err, "decode image file failed")
		return
	}

	// prepare BinaryBitmap
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		err = errors.Errorf(err, "gozxing NewBinaryBitmapFromImage failed ")
		return
	}

	// decode image
	barReader := oned.NewCode128Reader()
	result, err := barReader.Decode(bmp, nil)
	if err != nil {
		err = errors.Errorf(err, "gozxing/oned reader decode barcode128 image failed")
		return
	}

	return result.String(), nil
}
