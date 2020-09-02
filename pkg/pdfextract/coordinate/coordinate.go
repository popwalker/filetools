package coordinate

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"invtools/common"
	"invtools/pkg/util"

	"invtools/utils"

	"invtools/utils/errors"

	"github.com/gosuri/uiprogress"
	"github.com/skratchdot/open-golang/open"
)

const (
	cmdName = "coordinate"
)

type ExtractorCoordinate struct {
	input, output   string
	concurrency     int
	coordinateFlags []string
	coordinates     map[string]string
}

func NewExtractorCoordinate(input, output string, coordinates []string, concurrency int) *ExtractorCoordinate {
	return &ExtractorCoordinate{
		input:           input,
		output:          output,
		coordinateFlags: coordinates,
		concurrency:     concurrency,
	}
}

func (e *ExtractorCoordinate) Validate() error {
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

	if !(strings.HasSuffix(e.output, common.ExtCsv) || strings.HasSuffix(e.output, common.ExtExecl)) {
		return errors.Errorf(nil, "output 不是一个csv/xlsx文件, output:%s", e.output)
	}

	if err := e.formatCoordinateFlags(); err != nil {
		return errors.Errorf(err, "校验坐标格式失败")
	}
	return nil
}

func (e *ExtractorCoordinate) formatCoordinateFlags() error {

	m := make(map[string]string)
	for _, f := range e.coordinateFlags {
		arr := strings.Split(f, "=")
		if len(arr) != 2 {
			return errors.Errorf(nil, "解析坐标参数失败")
		}
		fname := arr[0]
		fvalue := arr[1]

		//strArr := strings.Fields(fvalue)
		if _, ok := m[fname]; !ok {
			m[fname] = fvalue
		}
	}

	e.coordinates = m

	return nil
}

func (e *ExtractorCoordinate) Extract() error {
	// validate
	if err := e.Validate(); err != nil {
		return errors.Errorf(err, "参数校验失败")
	}

	// execute
	return e.execute()
}

func (e *ExtractorCoordinate) execute() error {
	files, err := util.ReadDirFiles(e.input, common.ExtPDF)
	if err != nil {
		return errors.Errorf(err, "read input directory failed")
	}

	if len(files) == 0 {
		fmt.Printf("[%s] 扫描指定目录后，没有得到文件", cmdName)
		return nil
	}

	fmt.Printf("[%s] 扫描路径后一共得到%d个文件,即将开始解析操作...\n", cmdName, len(files))

	groups := util.DivideSliceIntoGroup(files, e.concurrency)
	//fmt.Println("groups:", groups,",len:", len(groups[0]));os.Exit(1)
	uiprogress.Start()
	var (
		wg sync.WaitGroup
		//waitTime     = time.Millisecond * 100
		successFiles []string
		failedFiles  []string
		st           = time.Now()
		queue        = make(chan map[string]string, len(files))
		num          int32
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
				//fmt.Println("processing file:", f)
				fields, err := e.extract(f)
				if err != nil {
					failedFiles = append(failedFiles, fmt.Sprintf("file:%s, err:%v", f, err))
				} else {
					successFiles = append(successFiles, f)
					fields["file_name"] = path.Base(f)
					//put into chan
					queue <- fields
					//queue <- map[string]string{"hello": f}
					atomic.AddInt32(&num, 1)
				}

			}
		})(&grp)
	}

	//go utils.HandlePanic(context.Background(), func() {
	//    defer wg.Done()
	//    f, err := os.OpenFile("./aa.txt", os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_APPEND, 0644)
	//    if err != nil {
	//        panic(err)
	//    }
	//    defer f.Close()
	//    for i := 0; i < int(num); i++ {
	//        v := <-queue
	//        b, err := json.Marshal(v)
	//        if err != nil {
	//            panic(err)
	//        }
	//        _, err = f.WriteString(string(b) + err.Error() + "adfadf")
	//        if err != nil {
	//            panic(err)
	//        }
	//        fmt.Println("解析结果:", v)
	//    }
	//})()

	wg.Wait()
	uiprogress.Stop()

	w := util.NewTableWriter(e.output)
	if err := w.DecideWriter(); err != nil {
		return errors.Errorf(err, "创建file writer 失败,文件:%s", e.output)
	}
	defer w.Close()

	var mkeys Mapkeys
	for i := 0; i < int(num); i++ {
		v := <-queue
		//fmt.Println("解析结果:", v)
		if len(mkeys) == 0 {
			mkeys, _ = getOrderdSliceFromMap(v)
			if err := w.WriteRecord(mkeys); err != nil {
				return errors.Errorf(err, "write file header failed")
			}
		}

		var record []string
		for _, k := range mkeys {
			if v, ok := v[k]; ok {
				record = append(record, v)
			}
		}

		if err := w.WriteRecord(record); err != nil {
			return errors.Errorf(err, "写入一行数据到output文件失败")
		}
	}

	fmt.Printf("[%s] 本次解析, 一共成功%d个pdf, 失败%d个, 总耗时:%s\n", cmdName, len(successFiles), len(failedFiles), time.Since(st))
	fmt.Printf("[%s] 输出文件存放路径: %s\n", cmdName, e.output)
	if len(failedFiles) > 0 {
		fmt.Printf("失败文件:%s", failedFiles)
	}

	open.Run(path.Dir(e.output))

	return nil
}

type Mapkeys []string

func (m Mapkeys) Less(i, j int) bool {
	if m[i] < m[j] {
		return true
	}
	return false
}

func (m Mapkeys) Len() int {
	return len(m)
}

func (m Mapkeys) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func getOrderdSliceFromMap(m map[string]string) (mkeys Mapkeys, values []string) {
	for k, v := range m {
		mkeys = append(mkeys, k)
		values = append(values, v)
	}

	sort.Sort(mkeys)

	return mkeys, values
}

func (e *ExtractorCoordinate) extract(filePath string) (map[string]string, error) {

	var m = make(map[string]string)
	for k, v := range e.coordinates {
		f, err := ioutil.TempFile(common.CurrentDir, "*.txt")
		if err != nil {
			return nil, errors.Errorf(err, "create tmp file failed")
		}
		f.Close()
		//fmt.Println("临时文件需要清理:", f.Name())
		defer os.Remove(f.Name())
		text, err := util.ExtractTextByCoordinate(filePath, v, f.Name())
		if err != nil {
			return nil, errors.Errorf(err, "从pdf中解析text失败")
		}
		if _, ok := m[k]; !ok {
			m[k] = text
		}
	}
	return m, nil
}
