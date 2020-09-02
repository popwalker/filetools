package compatible

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"invtools/common"
	"invtools/logger"
	"invtools/pkg/util"
	"invtools/pkg/util/xpdf"
	"invtools/utils"
	"invtools/utils/errors"

	"github.com/gosuri/uiprogress"
	"github.com/otiai10/gosseract"
	"github.com/skratchdot/open-golang/open"
)

const (
	coordinatePrefix = "coord_"
	regexpPrefix     = "reg_"
	tmpDirName       = "pdfextract_tmp"
)

type Extractor struct {
	inputDir, outputFile     string
	rawArgs                  []string
	concurrency, maxReadPage int
	withCoordinate, withOcr  bool
	coordinates, regexps     sync.Map
	resultKeys               []string // code,date
	tmpDir                   string   // 临时路径
	withCnf                  string
	withDebug                bool
	Config                   []*ExtractConfig
}

func init() {
	logger.InitLog()
}

// NewExtractor instance an new Extractor
func NewExtractor(inputDir, outputFile string, concurrency, maxReadPage int, withCoordinate, withOcr bool, rawArgs []string, withCnf string, withDebug bool) *Extractor {
	return &Extractor{
		inputDir:       inputDir,
		outputFile:     outputFile,
		rawArgs:        rawArgs,
		concurrency:    concurrency,
		withCoordinate: withCoordinate, // 是否使用坐标
		withOcr:        withOcr,        // 是否使用Ocr
		coordinates:    sync.Map{},     // 保存的是坐标信息
		regexps:        sync.Map{},     // 保存的是编译好的正则表达式
		resultKeys:     []string{},     // 结果集中的字段名
		maxReadPage:    maxReadPage,    // 每张pdf最多读取几页用于解析
		withCnf:        withCnf,
		withDebug:      withDebug,
	}
}

// Validate .
func (e *Extractor) Validate() error {
	if e == nil {
		return errors.Errorf(nil, "receiver is nil")
	}

	if ok := utils.CheckDirIsExist(e.inputDir); !ok {
		return errors.Errorf(nil, "input directory not exists")
	}

	if !(strings.HasSuffix(e.outputFile, common.ExtCsv) || strings.HasSuffix(e.outputFile, common.ExtExecl)) {
		return errors.Errorf(nil, "output 不是一个csv/xlsx文件, output:%s", e.outputFile)
	}

	err := e.initParseArgs()
	if err != nil {
		return errors.Errorf(err, "初始化解析参数失败")
	}

	return nil
}

// initParseArgs 初始化参数
func (e *Extractor) initParseArgs() error {
	if e.withCnf == "" && len(e.rawArgs) == 0 {
		return errors.Errorf(nil, "解析参数不能为空")
	}

	var coordArgs, regArgs []string
	for _, v := range e.rawArgs {
		if strings.HasPrefix(v, coordinatePrefix) {
			coordArgs = append(coordArgs, strings.TrimPrefix(v, coordinatePrefix))
		}
		if strings.HasPrefix(v, regexpPrefix) {
			regArgs = append(regArgs, strings.TrimPrefix(v, regexpPrefix))
		}
	}

	for _, v := range coordArgs {
		arr := strings.Split(v, "=")
		if len(arr) != 2 {
			return errors.Errorf(nil, "解析坐标参数失败")
		}
		fname := arr[0]
		fvalue := arr[1]

		e.coordinates.Store(fname, fvalue)
	}

	for _, v := range regArgs {
		arr := strings.Split(v, "=")
		if len(arr) != 2 {
			return errors.Errorf(nil, "解析正则参数失败")
		}
		fname := arr[0]
		fvalue := regexp.MustCompile(arr[1])

		e.regexps.Store(fname, fvalue)
	}

	// map去重
	var keys = make(map[string]struct{})
	e.coordinates.Range(func(k, v interface{}) bool {
		keys[k.(string)] = struct{}{}
		return true
	})
	e.regexps.Range(func(k, v interface{}) bool {
		keys[k.(string)] = struct{}{}
		return true
	})

	// 换成数组
	var resultKeys []string
	for k, _ := range keys {
		resultKeys = append(resultKeys, k)
	}

	// 排序
	sort.Slice(resultKeys, func(i, j int) bool {
		if resultKeys[i] < resultKeys[j] {
			return true
		}
		return false
	})
	e.resultKeys = resultKeys

	return nil
}

// RegisterCleanUpFunc 注册清理函数，用于在意外退出后，清理临时路径
func (e *Extractor) RegisterCleanUpFunc() {
	//创建监听退出chan
	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP:
				e.cleanTmpDir()
				os.Exit(1)
			default:
				// do nothing
			}
		}
	}()
}

func (e *Extractor) Extract() error {
	// validate
	if err := e.Validate(); err != nil {
		return errors.Errorf(err, "参数校验失败")
	}

	// 注册清理函数
	e.RegisterCleanUpFunc()

	// 正常退出，清除目录
	defer e.cleanTmpDir()

	// 使用配置文件进行解析
	if e.withCnf != "" {
		return e.executeWithConf()
	}
	// execute
	return e.execute()
}

type Result struct {
	Filename string
	Data     map[string]string
}

type Results []*Result

func (r Results) Len() int {
	return len(r)
}
func (r Results) Less(i, j int) bool {
	if r[i].Filename < r[j].Filename {
		return true
	}
	return false
}
func (r Results) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (e *Extractor) execute() error {
	files, err := util.ReadDirFilesV3(e.inputDir, common.ExtPDF, common.ExtPng)
	if err != nil {
		return errors.Errorf(err, "read input directory failed")
	}
	if len(files) == 0 {
		fmt.Printf("扫描指定目录后，没有得到文件")
		return nil
	}
	fmt.Printf("扫描路径后一共得到%d个文件,即将开始解析操作...\n", len(files))

	var (
		wg              sync.WaitGroup
		st              = time.Now()
		success, failed int
		resultm         sync.Map
	)

	groups := util.DivideSliceIntoGroup(files, e.concurrency)
	uiprogress.Start()
	wg.Add(len(groups))

	for i := 0; i < len(groups); i++ {
		var (
			ctx = context.Background()
			grp = groups[i]
			cnt = len(grp)
		)
		go utils.HandlePanicV2(ctx, func(i interface{}) {
			ggrp := *i.(*[]string)
			defer wg.Done()

			bar := uiprogress.AddBar(cnt).AppendCompleted().PrependElapsed()
			bar.PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("processing: %d/%d", b.Current(), cnt)
			})

			for bar.Incr() {
				f := ggrp[bar.Current()-1]
				data, err := e.extract(f)
				if err != nil {
					failed++
				} else {
					success++
				}
				resultm.Store(path.Base(f), data)
			}
		})(&grp)
	}

	wg.Wait()
	uiprogress.Stop()

	fmt.Printf("本次解析, 一共成功%d个pdf, 失败%d个, 总耗时:%s\n", success, failed, time.Since(st))

	// 得到所有的结果
	var results Results
	resultm.Range(func(key, value interface{}) bool {
		v, ok := value.(*Result)
		if !ok {
			return false
		}
		results = append(results, v)
		return true
	})

	if len(results) == 0 {
		return errors.Errorf(nil, "解析结果为空")
	}

	// 排序
	sort.Sort(results)

	fmt.Println("开始生成结果")
	// 准备生成输出文件
	w := util.NewTableWriter(e.outputFile)
	if err := w.DecideWriter(); err != nil {
		return errors.Errorf(err, "创建file writer 失败,文件:%s", e.outputFile)
	}
	defer w.Close()

	// 写入第一行文件头
	var header []string
	header = append(header, "file_name")
	header = append(header, e.resultKeys...)
	if err := w.WriteRecord(header); err != nil {
		return errors.Errorf(err, "write file header failed")
	}

	// 写入数据
	for i := 0; i < len(results); i++ {
		result := results[i]

		var record []string
		record = append(record, result.Filename)
		for _, headerKey := range e.resultKeys {
			if v, ok := result.Data[headerKey]; ok {
				record = append(record, v)
			} else {
				record = append(record, "")
			}
		}
		if err := w.WriteRecord(record); err != nil {
			return errors.Errorf(err, "写入一行数据到output文件失败")
		}
	}
	fmt.Printf("结果文件存放路径: %s\n", e.outputFile)

	open.Run(path.Dir(e.outputFile))
	return nil
}

func (e *Extractor) extract(filePath string) (*Result, error) {
	var (
		result = &Result{
			Filename: path.Base(filePath),
			Data:     map[string]string{},
		}
		err error
	)
	if e.withCoordinate {
		err = e.extractWithCoordinate(result, filePath)
		if err != nil {
			err = e.extractWithOcr(result, filePath)
			if err != nil {
				// todo: 处理error
			}
		}
	} else if e.withOcr {
		err = e.extractWithOcr(result, filePath)
		if err != nil {
			// todo: 处理error
		}
	}
	return result, nil
}

func (e *Extractor) extractWithCoordinate(result *Result, filePath string) error {
	var extractSuccess bool
	e.coordinates.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.(string)
		f, err := ioutil.TempFile(common.CurrentDir, "*.txt")
		if err != nil {
			return false
			// todo: 处理error
			//return nil, errors.Errorf(err, "create tmp file failed")
		}
		f.Close()
		//fmt.Println("临时文件需要清理:", f.Name())
		defer os.Remove(f.Name())
		text, err := util.ExtractTextByCoordinate(filePath, v, f.Name())
		if err != nil {
			return false
			// todo: 处理error
			//return nil, errors.Errorf(err, "从pdf中解析text失败")
		}
		if _, ok := result.Data[k]; !ok {
			result.Data[k] = text
			if strings.TrimSpace(text) != "" {
				extractSuccess = true
			}
		}
		return true
	})
	if !extractSuccess {
		return errors.Errorf(nil, "坐标解析失败")
	}

	return nil
}

func (e *Extractor) extractWithOcr(result *Result, filePath string) error {
	switch strings.ToLower(path.Ext(filePath)) {
	case common.ExtPDF:
		return e.extractWithOcrFromPDF(result, filePath)
	case common.ExtPng:
		return e.extractWithOcrFromPng(result, filePath)
	default:
		return errors.Errorf(nil, "未支持的文件扩展名:%s", filePath)
	}
}

// getTmpDir 获取临时路径名称
func (e *Extractor) getTmpDir() string {
	// {当前路径}/pdfextract_tmp_{时间戳}
	if e.tmpDir == "" {
		tmp := path.Join(common.CurrentDir, fmt.Sprintf("%s_%d", tmpDirName, time.Now().Unix()))
		e.tmpDir = tmp
	}

	return e.tmpDir
}

// cleanTmpDir 清理临时路径
func (e *Extractor) cleanTmpDir() {
	tmpDirname := e.tmpDir
	if strings.Contains(tmpDirname, "_tmp") || utils.CheckDirIsExist(tmpDirname) {
		utils.RmAll(tmpDirname)
	}
}

// extractWithOcrFromPDF 使用ocr从pdf读取信息
func (e *Extractor) extractWithOcrFromPDF(result *Result, filePath string) error {

	cleanFilename := strings.TrimSuffix(path.Base(filePath), path.Ext(filePath))
	cleanFilename = fmt.Sprintf("%x", md5.Sum([]byte(cleanFilename)))

	// 临时路径: 当前路径/pdfextract_compatiable_tmp/{md5(filename)}
	tmpdir := e.getTmpDir()
	tmpDir := path.Join(tmpdir, cleanFilename)
	utils.CheckAndMkDir(tmpDir)

	files, err := xpdf.PdfToPngV2(filePath, tmpDir, cleanFilename)
	if err != nil {
		return errors.Errorf(err, "pdf转图片失败")
	}
	defer utils.RmAll(tmpDir)

	client := gosseract.NewClient()
	defer client.Close()

	var maxReadPage = e.maxReadPage
	if maxReadPage == 0 {
		maxReadPage = len(files)
	}

	for i := 0; i < len(files); i++ {
		filePath := files[i]
		err := client.SetImage(filePath)
		if err != nil {
			return errors.Errorf(err, "SetImageFromBytes失败")
		}
		//err = client.SetLanguage("Chinese")
		//if err != nil {
		//	return errors.Errorf(err, "ocr设置语言失败")
		//}

		text, err := client.Text()
		if err != nil {
			return errors.Errorf(err, "读取图片中的文字失败")
		}

		// 正则解析
		m := e.extractTextWithRegexp(text)
		for kk, vv := range m {
			if kk != "" && vv != "" {
				result.Data[kk] = vv
			}
		}

		if i+1 > maxReadPage {
			continue
		}
	}

	return nil
}

// extractWithOcrFromPng 使用ocr从图片读取信息
func (e *Extractor) extractWithOcrFromPng(result *Result, filePath string) error {
	client := gosseract.NewClient()
	defer client.Close()

	err := client.SetImage(filePath)
	if err != nil {
		return errors.Errorf(err, "SetImageFromBytes失败")
	}

	text, err := client.Text()
	if err != nil {
		return errors.Errorf(err, "读取图片中的文字失败")
	}

	// 正则解析
	// 正则解析
	m := e.extractTextWithRegexp(text)
	for kk, vv := range m {
		if kk != "" && vv != "" {
			result.Data[kk] = vv
		}
	}

	return nil
}

// extractTextWithRegexp 正则解析文本
func (e *Extractor) extractTextWithRegexp(text string) map[string]string {

	var m = make(map[string]string)
	e.regexps.Range(func(key, value interface{}) bool {
		regRex, ok := value.(*regexp.Regexp)
		if !ok {
			return false
		}

		res := regRex.FindSubmatch([]byte(text))
		if len(res) == 2 {
			hintKey := key.(string)
			hintValue := util.StringPurify(string(res[1]))
			m[hintKey] = hintValue
			return true
		}

		return true
	})

	return m
}
