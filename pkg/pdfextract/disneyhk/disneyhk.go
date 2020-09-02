package disneyhk

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"invtools/common"
	"invtools/pkg/util"

	"invtools/utils"

	"invtools/utils/errors"

	"github.com/gosuri/uiprogress"
)

const (
	cmdName = "disneyhk"
)

type extractorDisneyHK struct {
	input, output string
	concurrency   int
	fields        []string
	actions       Actions
}

func NewExtractorDisneyHK(input, output string, concurrency int, fields ...string) *extractorDisneyHK {
	return &extractorDisneyHK{
		input:       input,
		output:      output,
		concurrency: concurrency,
		fields:      fields,
	}
}

// Validate does some validate things
func (e *extractorDisneyHK) Validate() error {
	f, err := os.Open(e.input)
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

	return nil
}

func (e *extractorDisneyHK) Extract() error {
	// validate
	if err := e.Validate(); err != nil {
		return errors.Errorf(err, "参数校验失败")
	}

	// set actions
	if err := e.SetActions(); err != nil {
		return errors.Errorf(err, "设置Actions失败")
	}

	// execute
	return e.execute()
}

func (e *extractorDisneyHK) execute() error {
	files, err := util.ReadDirFiles(e.input, common.ExtPDF)
	if err != nil {
		return errors.Errorf(err, "read input directory failed")
	}

	fmt.Printf("[%s] 扫描路径后一共得到%d个文件,即将开始解析操作...\n", cmdName, len(files))

	groups := util.DivideSliceIntoGroup(files, e.concurrency)

	uiprogress.Start()
	var (
		wg           sync.WaitGroup
		waitTime     = time.Millisecond * 100
		successFiles []string
		failedFiles  []string
		st           = time.Now()
		queue        = make(chan []string, len(files))
	)

	wg.Add(len(groups) + 1)
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
				time.Sleep(waitTime)
				f := grp[bar.Current()-1]
				fields, err := e.extract(f)
				if err != nil {
					failedFiles = append(failedFiles, fmt.Sprintf("file:%s, err:%v", f, err))
				} else {
					successFiles = append(successFiles, f)
				}
				// put into chan
				queue <- fields
			}
		})(&grp)
	}

	go utils.HandlePanic(context.Background(), func() {
		defer wg.Done()
		for v := range queue {
			fmt.Println("解析结果:", v)
		}
	})

	wg.Wait()
	uiprogress.Stop()

	fmt.Printf("[%s] 本次一共检测了%d个pdf，失败%d个,总耗时:%s\n", len(successFiles), len(failedFiles), time.Since(st))

	return nil
}

func (e *extractorDisneyHK) extract(filePath string) ([]string, error) {
	//unipdf := util.NewUniPdf()
	//pageContent, err:= unipdf.ExtractTextWithPages(filePath, "", 0)
	return nil, nil
}

func (e *extractorDisneyHK) SetActions() error {
	if len(e.fields) == 0 {
		e.actions = defaultActions
		return nil
	}

	actions, err := e.GetActions()
	if err != nil {
		return errors.Errorf(err, "Get Actions failed")
	}
	e.actions = actions

	return nil
}

// GetTasks 根据fields获取需要的Actions
func (e *extractorDisneyHK) GetActions() (Actions, error) {
	var actions Actions
	for _, field := range e.fields {
		switch {
		case strings.EqualFold(field, FieldDescSequentialNo):
			actions = append(actions, SequentialNoAction)
		case strings.EqualFold(field, FieldDescReservationCode):
			actions = append(actions, ReservationCodeAction)
		default:
			return nil, errors.Errorf(nil, "Unknown field:%s", field)

		}
	}

	return actions, nil
}
