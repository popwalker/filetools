package xpdf

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"invtools/utils"
	"invtools/utils/errors"
)

// PDF转图片，使用xpdf/pdftopng,返回图片文件二进制流
func PdfToPng(pdfFilePath string, outputDir string, prefix string) ([][]byte, error) {
	if !utils.CheckFileIsExist(pdfFilePath) {
		return nil, errors.Errorf(nil, "目标pdf文件不存在")
	}

	err := utils.CheckAndMkDir(outputDir)
	if err != nil {
		return nil, errors.Errorf(err, "检查并创建路径失败,outputDir:%s", outputDir)
	}
	defer utils.RmAll(outputDir)

	cmd := fmt.Sprintf("%s %s %s", "pdftopng", pdfFilePath, path.Join(outputDir, prefix))
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	err = exec.CommandContext(ctx, "sh", "-c", cmd).Run()
	if err != nil {
		return nil, errors.Errorf(err, "执行command exec pdf转图片失败")
	}

	fis, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return nil, errors.Errorf(err, "读取文件夹失败")
	}

	if len(fis) == 0 {
		return nil, errors.Errorf(nil, "读取到的文件数为0")
	}

	var (
		pages []string
		m     = make(map[string][]byte)
		bs    [][]byte
	)
	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), prefix) && strings.HasSuffix(fi.Name(), ".png") {
			imgName := strings.TrimPrefix(fi.Name(), prefix+"-")
			pages = append(pages, imgName)

			b, err := ioutil.ReadFile(path.Join(outputDir, fi.Name()))
			if err != nil {
				return nil, errors.Errorf(err, "读取图片数据失败")
			}
			m[imgName] = b
		}
	}

	sort.Strings(pages)

	for _, v := range pages {
		if v, ok := m[v]; ok {
			bs = append(bs, v)
		}
	}

	return bs, nil
}

func PdfToPngV2(pdfFilePath string, outputDir string, prefix string) ([]string, error) {
	if !utils.CheckFileIsExist(pdfFilePath) {
		return nil, errors.Errorf(nil, "目标pdf文件不存在")
	}

	err := utils.CheckAndMkDir(outputDir)
	if err != nil {
		return nil, errors.Errorf(err, "检查并创建路径失败,outputDir:%s", outputDir)
	}

	pdfFilePath = strings.NewReplacer(" ", "\\ ", "(", "\\(", ")", "\\)", "（", "\\（", "）", "\\）").Replace(pdfFilePath)

	cmd := fmt.Sprintf("%s %s %s", "pdftopng", pdfFilePath, path.Join(outputDir, prefix))
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	err = exec.CommandContext(ctx, "sh", "-c", cmd).Run()
	if err != nil {
		return nil, errors.Errorf(err, "执行command exec pdf转图片失败")
	}

	fis, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return nil, errors.Errorf(err, "读取文件夹失败")
	}

	if len(fis) == 0 {
		return nil, errors.Errorf(nil, "读取到的文件数为0")
	}

	var (
		pages []string
		m     = make(map[string]string)
		files []string
	)
	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), prefix) && strings.HasSuffix(fi.Name(), ".png") {
			imgName := strings.TrimPrefix(fi.Name(), prefix+"-")
			pages = append(pages, imgName)
			m[imgName] = path.Join(outputDir, fi.Name())
		}
	}

	sort.Strings(pages)

	for _, v := range pages {
		if v, ok := m[v]; ok {
			files = append(files, v)
		}
	}

	return files, nil
}

func PdfToPngBytes(pdfFilePath, outputDir, prefix string) ([][]byte, error) {
	if !utils.CheckFileIsExist(pdfFilePath) {
		return nil, errors.Errorf(nil, "目标pdf文件不存在")
	}

	err := utils.CheckAndMkDir(outputDir)
	if err != nil {
		return nil, errors.Errorf(err, "检查并创建路径失败,outputDir:%s", outputDir)
	}

	pdfFilePath = strings.NewReplacer(" ", "\\ ", "(", "\\(", ")", "\\)", "（", "\\（", "）", "\\）").Replace(pdfFilePath)

	cmd := fmt.Sprintf("%s %s %s", "pdftopng", pdfFilePath, path.Join(outputDir, prefix))
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	err = exec.CommandContext(ctx, "sh", "-c", cmd).Run()
	if err != nil {
		return nil, errors.Errorf(err, "执行command exec pdf转图片失败")
	}

	fis, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return nil, errors.Errorf(err, "读取文件夹失败")
	}

	if len(fis) == 0 {
		return nil, errors.Errorf(nil, "读取到的文件数为0")
	}

	var imagesBytes [][]byte
	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), prefix) && strings.HasSuffix(fi.Name(), ".png") {

			b, err := ioutil.ReadFile(path.Join(outputDir, fi.Name()))
			if err != nil {
				return nil, errors.Errorf(err, "读取png文件失败")
			}
			imagesBytes = append(imagesBytes, b)

		}
	}

	return imagesBytes, nil
}
