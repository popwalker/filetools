package pdfrepair

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"invtools/pkg/util/mupdf"

	"invtools/utils"

	"invtools/utils/errors"

	"github.com/skratchdot/open-golang/open"
)

const (
	cmdName = "pdfrepair"
)

type PdfRepair struct {
	inputDir, outputDir string
}

func NewPdfRepair(inputDir, outputDir string) *PdfRepair {
	return &PdfRepair{
		inputDir:  inputDir,
		outputDir: outputDir,
	}
}

// Validate .
func (r *PdfRepair) Validate() error {
	if r == nil {
		return errors.Errorf(nil, "receiver is nil")
	}

	if ok := utils.CheckDirIsExist(r.inputDir); !ok {
		return errors.Errorf(nil, "inputDir not exists")
	}

	if ok := utils.CheckDirIsExist(r.outputDir); !ok {
		err := utils.Mkdir(r.outputDir)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (r *PdfRepair) Do() error {
	if err := r.Validate(); err != nil {
		return errors.Errorf(err, "校验失败")
	}

	defer func() {
		var hasFile bool
		fis, err := ioutil.ReadDir(r.outputDir)
		if err != nil {
			return
		}
		for i := 0; i < len(fis); i++ {
			fi := fis[i]
			if strings.HasSuffix(strings.ToLower(fi.Name()), ".pdf") {
				hasFile = true
				break
			}
		}
		if !hasFile {
			utils.RmAll(r.outputDir)
		}
	}()

	var files []string // 文件路径集合
	err := filepath.Walk(r.inputDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".pdf") {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return errors.Errorf(err, "读取目标目录出错")
	}

	if len(files) == 0 {
		fmt.Printf("[%s] 目录: %s 下没有pdf文件\n", cmdName, r.inputDir)
		return nil
	}

	for i := 0; i < len(files); i++ {
		infile := files[i]
		outfile := path.Join(r.outputDir, path.Base(infile))
		fmt.Printf("[%s] [%d/%d]start fix file: %s\n", cmdName, i+1, len(files), path.Base(infile))
		err := mupdf.PdfRepair(infile, outfile)
		if err != nil {
			fmt.Printf("[%s] 修复文件:%s发生错误,err:%s\n", cmdName, path.Base(infile), err.Error())
			continue
		}
	}

	fmt.Printf("\n[%s] 批量修复完毕, 修复后的文件保存目录是: %s\n", cmdName, r.outputDir)
	open.Run(path.Dir(r.outputDir))
	return nil
}
