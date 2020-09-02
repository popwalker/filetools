package pdfsplit

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"invtools/common"
	"invtools/pkg/util"

	"invtools/utils"

	"invtools/utils/errors"

	"github.com/gosuri/uiprogress"
	rscPdf "github.com/rsc.io/pdf"
	"github.com/skratchdot/open-golang/open"
	"github.com/unidoc/unidoc/common/license"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

const (
	cmdName = "pdfsplit"
)

type Splitter struct {
	input, output, password string
	perPage, concurrency    int
}

func init() {
	const unidocLicenseKey = `
-----BEGIN UNIDOC LICENSE KEY-----
eyJsaWNlbnNlX2lkIjoiYjIxYTQzOWQtM2NmYS00NmVjLTRjZmUtYTQ1NzkwMjY2NDEwIiwiY3VzdG9tZXJfaWQiOiIxM2VmZDM1MS1mYmQxLTRlNDctNzUzZS1jMzZlZWEzNzVlYWQiLCJjdXN0b21lcl9uYW1lIjoiS2xvb2sgVHJhdiIsImN1c3RvbWVyX2VtYWlsIjoiaXRAa2xvb2suY29tIiwidGllciI6ImJ1c2luZXNzIiwiY3JlYXRlZF9hdCI6MTU0NTk4OTA0MSwiZXhwaXJlc19hdCI6MCwiY3JlYXRvcl9uYW1lIjoiVW5pRG9jIFN1cHBvcnQiLCJjcmVhdG9yX2VtYWlsIjoic3VwcG9ydEB1bmlkb2MuaW8ifQ==
+
PcQ6I/T42jVYx+3wIRqLw1sbj4sXIhTnZ/plDfNy1Njyp+6Cw4ria/K5NW0qKJHQeKBN3yJbK2siHFHiOnmtVkx74SI193YgnUKSwDkPncCkLMr4cdko64rb4PG7ClrU85dTCcqw5KVrlod+pVtIjQhLYfkmCBnJ3geGKBAN7qkT+ImktL94p9iS/gRqi3Dj02YLIuBKyHoDJBWKvc2Ae3duSvyLWt1LckMT1qfOfyWH24D6KCdPodE/YzEyaYKkkatpMvznfWhqQmkbNZeMqMuLkmjNMgBElF+6hxKbzS6NsY0HNsm2jM4h6iclN7E5zMDnHS5hOpPPjEHE4fQiBg==
-----END UNIDOC LICENSE KEY-----`

	err := license.SetLicenseKey(unidocLicenseKey)
	if err != nil {
		fmt.Println("PDF Error loading license, err:", err)
		os.Exit(1)
	}
}

func NewPdfSplitter(input, output, password string, perPage, concurrency int) *Splitter {
	return &Splitter{
		input:       input,
		output:      output,
		password:    password,
		perPage:     perPage,
		concurrency: concurrency,
	}
}

func (s *Splitter) validate() error {
	f, err := os.Open(s.input)
	if err != nil {
		return errors.Errorf(err, "open input directory failed")
	}
	defer f.Close()

	fi, err := os.Stat(f.Name())
	if err != nil {
		return errors.Errorf(err, "cannot stat input directory")
	}

	if !fi.IsDir() {
		return errors.Errorf(err, "input 不是一个目录,请检查")
	}

	if err := utils.CheckAndMkDir(s.output); err != nil {
		return errors.Errorf(err, "创建目录失败,目录:%s", s.output)
	}

	return nil
}

func (s *Splitter) Do() error {
	if err := s.validate(); err != nil {
		return err
	}

	return s.execute()
}

func (s *Splitter) execute() error {
	files, err := util.ReadDirFiles(s.input, common.ExtPDF)
	if err != nil {
		return errors.Errorf(err, "读取input路径下的pdf文件失败")
	}

	fmt.Printf("[%s] 扫描路径后一共得到%d个文件,即将开始拆分操作...\n", cmdName, len(files))

	groups := util.DivideSliceIntoGroup(files, s.concurrency)

	uiprogress.Start()
	var (
		wg sync.WaitGroup
		//waitTime     = time.Millisecond * 100
		st           = time.Now()
		successFiles []string
		failedFiles  []string
	)
	//atomic.StoreInt32(&num, 0)

	wg.Add(len(groups))
	//wg.Add(len(groups) + 1)
	for _, value := range groups {
		ctx := context.Background()
		grp := value
		cnt := len(grp)
		go utils.HandlePanicV2(ctx, func(i interface{}) {
			grp := *i.(*[]string)
			defer wg.Done()

			bar := uiprogress.AddBar(cnt).AppendCompleted().PrependElapsed()
			bar.PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("processing: %d/%d", b.Current(), cnt)
			})

			for bar.Incr() {
				//time.Sleep(waitTime)
				f := grp[bar.Current()-1]
				err := s.split(f)
				if err != nil {
					failedFiles = append(failedFiles, fmt.Sprintf("file:%s, err:%v", f, err))
				} else {
					successFiles = append(successFiles, f)
				}
			}
		})(&grp)
	}

	wg.Wait()
	uiprogress.Stop()

	fmt.Printf("[%s] 本次拆分操作, 一共成功%d个pdf, 失败%d个, 总耗时:%s\n", cmdName, len(successFiles), len(failedFiles), time.Since(st))
	fmt.Printf("[%s] 拆分后的文件保存目录是: %s\n", cmdName, s.output)
	if len(failedFiles) > 0 {
		fmt.Printf("失败文件:%s", failedFiles)
	}

	open.Run(path.Dir(s.output))

	return nil
}

func (s *Splitter) split(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return errors.Errorf(err, "open file failed, file:%s", filePath)
	}
	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return err
	}

	if isEncrypted {
		if s.password == "" {
			_, err = pdfReader.Decrypt([]byte(""))
		} else {
			_, err = pdfReader.Decrypt([]byte(s.password))
		}
		if err != nil {
			return errors.Errorf(err, "解密失败")
		}
	}

	fi, err := f.Stat()
	if err != nil {
		return errors.Errorf(err, "get file stat failed")
	}

	unidocNumPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	// 使用rsc.io/pdf读取文件
	rscReader, err := rscPdf.NewReader(f, fi.Size())
	if err != nil {
		return err
	}

	// 获取页数
	rscNumPages := rscReader.NumPage()

	// 默认页数以unidoc读取的为准，如果和rscpdf读取的不一致，则以rscpdf读取的页数为准
	numPages := unidocNumPages
	if rscNumPages != unidocNumPages {
		numPages = rscNumPages
	}

	if numPages%s.perPage != 0 {
		return errors.Errorf(nil, "文件只有%d页，无法按照平均每%d页进行拆分,file:%s", numPages)
	}

	n := 0
	for i := 1; i <= numPages; i += s.perPage {
		pageFrom, pageTo := i, i-1+s.perPage

		pdfWriter := pdf.NewPdfWriter()
		n++
		for i := pageFrom; i <= pageTo; i++ {
			pageNum := i

			page, err := pdfReader.GetPage(pageNum)
			if err != nil {
				return err
			}

			err = pdfWriter.AddPage(page)
			if err != nil {
				return err
			}
		}
		fWrite, err := os.Create(getSubFileName(s.output, filePath, n))
		if err != nil {
			return err
		}

		err = pdfWriter.Write(fWrite)
		if err != nil {
			fWrite.Close()
			return err
		}
		fWrite.Close()
	}

	return nil
}

func getSubFileName(dir, filePath string, n int) string {
	fileName := path.Base(filePath)
	ext := path.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	return fmt.Sprintf("%s/%s_%d%s", dir, name, n, ext)
}
