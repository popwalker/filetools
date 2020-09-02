package pdfcpu

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"invtools/pkg/util"
	"invtools/utils"
	"invtools/utils/errors"
)

// PdfToPng 使用pdfcpu进行图片解析
func PdfToPng(pdfFilePath string, outputDir string) ([]string, error) {
	if !utils.CheckFileIsExist(pdfFilePath) {
		return nil, errors.Errorf(nil, "目标pdf文件不存在")
	}

	cmd := fmt.Sprintf("pdfcpu extract -mode image '%s' %s", pdfFilePath, outputDir)
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	err := exec.CommandContext(ctx, "sh", "-c", cmd).Run()
	if err != nil {
		return nil, errors.Errorf(err, "执行command exec pdfcpu解析图片出错, file:[%s]", pdfFilePath)
	}

	files, err := util.ReadDirFilesV3(outputDir, "png")
	if err != nil {
		return nil, errors.Errorf(err, "读取pdfcpu解析出来的图片文件出错")
	}

	return files, nil
}
