package util

import (
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path"

	"invtools/common"
	"invtools/utils"
	"invtools/utils/errors"

	"github.com/disintegration/gift"
)

// CropPdfToImage 裁切图片
func CropPdfToImage(imageFilePath string, coordinates []int, dstDir string) (string, error) {
	if !utils.CheckFileIsExist(imageFilePath) {
		return "", errors.Errorf(nil, "imageFilePath file not exists:%s", imageFilePath)
	}

	if !utils.CheckDirIsExist(dstDir) {
		return "", errors.Errorf(nil, "dir not exists:%s", dstDir)
	}

	filter := gift.Crop(image.Rectangle{
		Min: image.Point{coordinates[0], coordinates[1]},
		Max: image.Point{coordinates[2], coordinates[3]},
	})

	src, err := loadImage(imageFilePath)
	if err != nil {
		return "", errors.Errorf(err, "加载图片失败")
	}

	g := gift.New(filter)
	dst := image.NewNRGBA(g.Bounds(src.Bounds()))
	g.Draw(dst, src)

	dstFilePath := path.Join(dstDir, GetPureFileName(imageFilePath)) + ".png"

	err = saveImage(dstFilePath, dst)
	if err != nil {
		return "", errors.Errorf(err, "保存切割后的文件失败")
	}
	return dstFilePath, nil
}

func loadImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Errorf(err, "os.Open file failed")
	}
	defer f.Close()

	ext := path.Ext(filename)
	switch ext {
	case common.ExtPng:
		return png.Decode(f)
	case common.ExtJpeg, common.ExtJpg:
		return jpeg.Decode(f)
	default:
		return nil, errors.Errorf(nil, "unsupported image file ext")
	}
}

func saveImage(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Errorf(err, "os.Create:%s failed", filename)
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		return errors.Errorf(err, "png.Encode failed, file:%s", filename)
	}
	return nil
}
