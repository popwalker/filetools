package qrscan

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
	"github.com/skratchdot/open-golang/open"
)

const (
	cmdName = "qrscanner"
)

type QrScanner struct {
	input, output, qrType string
	concurrency           int
}

func NewQrScanner(input, output, qrType string, concurrency int) *QrScanner {
	return &QrScanner{
		input:       input,
		output:      output,
		qrType:      qrType,
		concurrency: concurrency,
	}
}

func (qs *QrScanner) Validate() error {
	f, err := os.Open(qs.input)
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

	outputFileExt := path.Ext(qs.output)
	if !common.AllowedOutputTable(outputFileExt) {
		return errors.Errorf(nil, "output文件类型:%s不支持", outputFileExt)
	}

	return nil
}

func (qs *QrScanner) Do() error {
	if err := qs.Validate(); err != nil {
		return err
	}

	return qs.execute()
}

type qrInfo struct {
	Filename string
	Code     string
}

func (qs *QrScanner) execute() error {
	files, err := util.ReadDirFilesV2(qs.input)
	if err != nil {
		return errors.Errorf(err, "读取input路径下的pdf文件失败")
	}

	if len(files) == 0 {
		fmt.Printf("[%s] 扫描指定目录后，没有得到文件", cmdName)
		return nil
	}

	fmt.Printf("[%s] 扫描路径后一共得到%d个文件,即将开始解析操作...\n", cmdName, len(files))

	groups := util.DivideSliceIntoGroup(files, qs.concurrency)

	uiprogress.Start()
	var (
		wg sync.WaitGroup
		//waitTime     = time.Millisecond * 100
		st           = time.Now()
		successFiles []string
		failedFiles  []string
		//ch           = make(chan *qrInfo, len(files))
		m = &sync.Map{}
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
				err := qs.scan2(m, f)
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

	fmt.Printf("[%s] 本次code解析操作, 一共成功解析%d个图片, 失败%d个, 总耗时:%s\n", cmdName, len(successFiles), len(failedFiles), time.Since(st))
	if len(failedFiles) > 0 {
		fmt.Printf("失败文件:%s", failedFiles)
	}

	//if len(ch) > 0 {
	fmt.Printf("[%s] 开始生成输出文件:%s\n", cmdName, qs.output)

	if err := qs.out2(m); err != nil {
		return errors.Errorf(err, "生成输出文件失败")
	}

	fmt.Printf("[%s] 解析完成, 输出文件路径是: %s\n", cmdName, qs.output)
	open.Run(path.Dir(qs.output))
	//}

	return nil
}

func (qs *QrScanner) scan(ch chan *qrInfo, filePath string) error {
	var (
		code string
		err  error
	)

	switch {
	case strings.EqualFold(qs.qrType, common.CodeTypeQRCode):
		code, err = util.QrCodeScan(filePath)
	case strings.EqualFold(qs.qrType, common.CodeTypeBarcode128):
		code, err = util.Barcode128Scan(filePath)
	default:
		err = errors.Errorf(nil, "暂不支持的code类型:%s", qs.qrType)
	}

	if err != nil {
		return errors.Errorf(err, "扫描解析code图片失败")
	}

	ch <- &qrInfo{
		Filename: path.Base(filePath),
		Code:     code,
	}

	return nil
}

func (qs *QrScanner) scan2(m *sync.Map, filePath string) error {
	var (
		code string
		err  error
	)

	switch {
	case strings.EqualFold(qs.qrType, common.CodeTypeQRCode):
		code, err = util.QrCodeScan(filePath)
	case strings.EqualFold(qs.qrType, common.CodeTypeBarcode128):
		code, err = util.Barcode128Scan(filePath)
	default:
		err = errors.Errorf(nil, "暂不支持的code类型:%s", qs.qrType)
	}

	if err != nil {
		return errors.Errorf(err, "扫描解析code图片失败")
	}

	m.Store(path.Base(filePath), &qrInfo{
		Filename: path.Base(filePath),
		Code:     code,
	})

	return nil
}

func (qs *QrScanner) out(ch chan *qrInfo) error {
	defer close(ch)

	w := util.NewTableWriter(qs.output)
	if err := w.DecideWriter(); err != nil {
		return errors.Errorf(err, "创建file writer 失败,文件:%s", qs.output)
	}
	defer w.Close()

	headerWrote := false
	l := len(ch)
	for i := 0; i < l; i++ {
		info := <-ch
		if !headerWrote {
			if err := w.WriteRecord([]string{"file_name", "code"}); err != nil {
				return errors.Errorf(err, "写入文件头失败")
			}
			headerWrote = true
		}

		if err := w.WriteRecord([]string{info.Filename, info.Code}); err != nil {
			return errors.Errorf(err, "写入一行数据到output文件失败")
		}
	}

	return nil
}

func (qs *QrScanner) out2(m *sync.Map) error {
	w := util.NewTableWriter(qs.output)
	if err := w.DecideWriter(); err != nil {
		return errors.Errorf(err, "创建file writer 失败,文件:%s", qs.output)
	}
	defer w.Close()

	headerWrote := false
	m.Range(func(k, v interface{}) bool {
		info := v.(*qrInfo)
		if !headerWrote {
			if err := w.WriteRecord([]string{"file_name", "code"}); err != nil {
				return false
			}
			headerWrote = true
		}

		if err := w.WriteRecord([]string{info.Filename, info.Code}); err != nil {
			return false
		}
		return true
	})

	return nil
}
