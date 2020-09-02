package xpdf

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"invtools/utils"
	"invtools/utils/errors"
)

// PDF转图片，使用xpdf/pdftotext,返回pdf文本
func PdfToText(pdfFilePath string, outputFilePath string) (string, error) {
	if !utils.CheckFileIsExist(pdfFilePath) {
		return "", errors.Errorf(nil, "目标pdf文件不存在")
	}

	cmd := fmt.Sprintf("pdftotext -enc UTF-8 -simple '%s' '%s'", pdfFilePath, outputFilePath)
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	err := exec.CommandContext(ctx, "sh", "-c", cmd).Run()
	if err != nil {
		return "", errors.Errorf(err, "执行command exec pdf转文字失败, file:[%s]", pdfFilePath)
	}
	defer func() {
		os.Remove(outputFilePath)
	}()

	b, err := ioutil.ReadFile(outputFilePath)
	if err != nil {
		return "", errors.Errorf(err, "pdf转文字后读取文件失败")
	}

	return string(b), nil
}
