package compatible

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"invtools/common"
	"invtools/logger"
	"invtools/pkg/util"
	"invtools/pkg/util/pdfcpu"
	"invtools/pkg/util/xpdf"
	"invtools/utils/errors"

	"invtools/utils"

	"github.com/gosuri/uiprogress"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/otiai10/gosseract"
	"github.com/skratchdot/open-golang/open"
)

func (e *Extractor) executeWithConf() error {
	err := e.parseConf()
	if err != nil {
		return errors.Errorf(err, "获取配置文件失败")
	}

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
				data, err := e.extractWithConf(f)
				if err != nil {
					if e.withDebug {
						logger.LoggerSugar.Errorf("extractWithConf err:%s", err)
					}
					failed++
				} else {
					resultm.Store(path.Base(f), data)
					success++
				}

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
	outputFilePath := getOutputFileFromRcv(e.inputDir, e.outputFile)
	// 准备生成输出文件
	w := util.NewTableWriter(outputFilePath)
	if err := w.DecideWriter(); err != nil {
		return errors.Errorf(err, "创建file writer 失败,文件:%s", e.outputFile)
	}
	defer w.Close()

	// 写入第一行文件头
	var header []string
	var fields []string
	header = append(header, "file_name")
	for _, cnf := range e.Config {
		header = append(header, cnf.FieldName)
		fields = append(fields, cnf.FieldName)
	}

	if err := w.WriteRecord(header); err != nil {
		return errors.Errorf(err, "write file header failed")
	}

	// 写入数据
	for i := 0; i < len(results); i++ {
		result := results[i]

		var record []string
		record = append(record, result.Filename)
		for _, fieldName := range fields {
			if v, ok := result.Data[fieldName]; ok {
				record = append(record, v)
			} else {
				record = append(record, "")
			}
		}
		if err := w.WriteRecord(record); err != nil {
			return errors.Errorf(err, "写入一行数据到output文件失败")
		}
	}
	fmt.Printf("结果文件存放路径: %s\n", outputFilePath)

	open.Run(path.Dir(e.outputFile))
	return nil
}

const customFlag = "system"

func getOutputFileFromRcv(originInputFilename, originOutputFilename string) string {
	if !strings.Contains(path.Base(originOutputFilename), customFlag) {
		return originOutputFilename
	}

	pieces := strings.Split(originInputFilename, "/")
	for _, v := range pieces {
		if strings.HasPrefix(v, "RCV") {
			return path.Join(path.Dir(originOutputFilename), fmt.Sprintf("%s%s", v, path.Ext(originOutputFilename)))
		}
	}
	return originOutputFilename

}

const (
	txtToolUnipdf = "unipdf"
	txtToolXpdf   = "pdftotext"
	txtToolOcr    = "ocr"
)

var defaultTextExtractTools = []string{txtToolXpdf, txtToolUnipdf, txtToolOcr}

type ExtractConfig struct {
	FieldName string `json:"field_name"`
	PageNum   int    `json:"page_num"`
	// tet,正则匹配文字(ocr转文字/pdf转文字)，条码扫描(pdf解析出图片/图片切割)
	// 枚举值: tet/reg/scan
	ExtractMethod    string   `json:"extract_method"`
	TextExtractTools []string `json:"text_extract_tool"` // [unipdf,pdftotext,ocr]
	TetCoordinates   []string `json:"tet_coordinates"`
	CropCoordinates  []int    `json:"crop_coordinates"` // [minX, minY, maxX, maxY]
	RegExp           string   `json:"reg_exp"`
	compiledRegExp   *regexp.Regexp            // 编译后的正则表达式
	CodeType         string `json:"code_type"` // qrcode, barcode128
}

func (e *Extractor) parseConf() error {
	if e.Config != nil {
		return nil
	}
	if !utils.CheckFileIsExist(e.withCnf) {
		return errors.Errorf(nil, "config file not exists")
	}
	b, err := ioutil.ReadFile(e.withCnf)
	if err != nil {
		return errors.Errorf(err, "read config file error,file:%s", e.withCnf)
	}

	var cnf []*ExtractConfig
	err = json.Unmarshal(b, &cnf)
	if err != nil {
		return errors.Errorf(err, "unmarshal config data failed")
	}

	e.Config = cnf
	return nil
}

const (
	ExtractMethodTET  = "tet"  // 使用tet坐标匹配
	ExtractMethodReg  = "reg"  // 使用正则匹配：解析text进行匹配/ocr转文字进行匹配
	ExtractMethodScan = "scan" // 使用扫描：解析图片进行扫描/切割图片进行扫描
)

func (e *Extractor) extractWithConf(filePath string) (*Result, error) {
	var (
		result = &Result{
			Filename: path.Base(filePath),
			Data:     map[string]string{},
		}
	)

	if e.Config == nil {
		return nil, errors.Errorf(nil, "配置文件为空")
	}
	se := SingleFileExtractor{
		filePath:  filePath,
		extractor: e,
	}
	if err := se.initPageResource(); err != nil {
		return nil, errors.Errorf(err, "初始化pageResource失败")
	}

	// 清除临时路径
	defer func() {
		utils.RmAll(se.tmpDir)
	}()

	for i := 0; i < len(e.Config); i++ {
		cnf := e.Config[i]
		var (
			err   error
			value string
		)
		switch strings.ToLower(cnf.ExtractMethod) {
		case ExtractMethodTET:
			value, err = se.extractWithTET(cnf)
		case ExtractMethodReg:
			value, err = se.extractWithRegV2(cnf)
		case ExtractMethodScan:
			value, err = se.extractWithScan(cnf)
		default:
			return nil, errors.Errorf(nil, "配置项中的ExtractMethod不合法")
		}
		if err != nil {
			return nil, errors.Errorf(err, "解析过程出错")
		}

		result.Data[cnf.FieldName] = value
	}

	return result, nil
}

// SingleFileExtractor 单个文件解析器
type SingleFileExtractor struct {
	filePath   string
	tmpDir     string
	tmpPDFDir  string // 每个文件一个专属临时保存pdf的目录
	tmpPngDir  string // 每个文件一个专属临时保存png的目录,ocr转换的图片
	tmpJpegDir string // unidoc解析出来的图片保存位置
	PageNumber int    // 页数
	extractor  *Extractor
	resource   map[int]*PageResource
}

// PageResource 每一页pdf的资源
type PageResource struct {
	filePath            string   // 拆分后的pdf文件保存位置
	ocrImage            []byte   // ocr扫描得到的图片
	ocrText             string   // ocr解析出来的文字
	extractedText       string   // unidoc解析出来的文字
	extractedImages     [][]byte // unidoc解析出来的图片
	extractedImageFiles []string // unidoc解析出来的图片文件
	croppedImages       [][]byte // 切割出来的图片
}

func (se *SingleFileExtractor) initPageResource() error {

	var splittedFiles []string
	var err error
	if !isNeedSplit(se.extractor.Config) {
		splittedFiles = []string{se.filePath}
	} else {
		cleanFilename := strings.TrimSuffix(path.Base(se.filePath), path.Ext(se.filePath))
		cleanFilename = fmt.Sprintf("%x", md5.Sum([]byte(cleanFilename)))
		// 临时路径: 当前路径/pdfextract_compatiable_tmp/{md5(filename)}
		tmpdir := se.extractor.getTmpDir()
		tmpDir := path.Join(tmpdir, cleanFilename)
		utils.CheckAndMkDir(tmpDir)
		se.tmpPDFDir = tmpDir

		splittedFiles, err = util.NewUniPdf().SplitIntoFiles(se.filePath, se.tmpPDFDir)
		if err != nil {
			return errors.Errorf(err, "拆分pdf失败")
		}
	}

	for i := 0; i < len(splittedFiles); i++ {
		if se.resource == nil {
			se.resource = make(map[int]*PageResource)
		}
		se.resource[i+1] = &PageResource{
			filePath: splittedFiles[i],
		}
	}

	cleanFilename := utils.GetUUIDString()
	// 临时路径: 当前路径/pdfextract_compatiable_{时间戳}
	tmpdir := se.extractor.getTmpDir()

	se.tmpDir = path.Join(tmpdir, cleanFilename)

	tmpPDFDir := path.Join(se.tmpDir, "tmp_pdf")
	utils.CheckAndMkDir(tmpPDFDir)
	se.tmpPDFDir = tmpPDFDir

	tmpPNGDir := path.Join(se.tmpDir, "tmp_png")
	utils.CheckAndMkDir(tmpPNGDir)
	se.tmpPngDir = tmpPNGDir

	tmpJPEGDir := path.Join(se.tmpDir, "tmp_jpeg")
	utils.CheckAndMkDir(tmpJPEGDir)
	se.tmpJpegDir = tmpJPEGDir

	return nil
}

func (se *SingleFileExtractor) extractWithScan(cnf *ExtractConfig) (string, error) {
	resource, ok := se.resource[cnf.PageNum]
	if !ok {
		return "", errors.Errorf(nil, "配置中的页码不存在,cnf.PageNum:%d, file:%s", cnf.PageNum, se.filePath)
	}

	if resource.extractedImageFiles == nil {
		// 先从pdf中解析出图片资源
		imagesFiles, err := util.NewUniPdf().ExtractImagesIntoFiles(resource.filePath, se.tmpJpegDir)
		if err != nil {
			imagesFiles, err = pdfcpu.PdfToPng(resource.filePath, se.tmpJpegDir)
			if err != nil {
				logger.Errorf("使用pdfcpu解析图片也出错,err:%+v", err)
			}
		}
		resource.extractedImageFiles = imagesFiles
	}

	for i := 0; i < len(resource.extractedImageFiles); i++ {
		imgFile := resource.extractedImageFiles[i]
		f, err := os.Open(imgFile)
		if err != nil {
			return "", errors.Errorf(err, "读取jpeg文件失败")
		}

		img, err := util.ImgDecode(path.Ext(imgFile), f)
		if err != nil {
			return "", errors.Errorf(err, "jpeg.Decode failed")
		}

		// prepare BinaryBitmap
		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			return "", errors.Errorf(err, "gozxing NewBinaryBitmapFromImage failed ")
		}

		var gozxingReader gozxing.Reader
		switch cnf.CodeType {
		case common.CodeTypeQRCode:
			// decode image
			gozxingReader = qrcode.NewQRCodeReader()
		case common.CodeTypeBarcode128:
			gozxingReader = oned.NewCode128Reader()
		default:
			return "", errors.Errorf(nil, "为支持的code类型:[%s]", cnf.CodeType)
		}
		result, err := gozxingReader.Decode(bmp, nil)
		if err != nil {
			// 出错继续
			continue
		}

		// 匹配到则返回
		return result.String(), nil
	}

	if len(cnf.CropCoordinates) == 0 {
		return "", nil
	}
	// todo: 解析图片没成功的话，使用图片切割
	pngfiles, err := xpdf.PdfToPngV2(resource.filePath, se.tmpPngDir, utils.GetRequestID())
	if err != nil {
		return "", errors.Errorf(err, "pdf转图片失败")
	}

	if len(pngfiles) < cnf.PageNum {
		return "", errors.Errorf(nil, "pdftopng转换后的图片数量:%d,配置中的PageNum:%d", len(pngfiles), cnf.PageNum)
	}

	srcPngFile := pngfiles[cnf.PageNum-1]
	croppedPngFile, err := util.CropPdfToImage(srcPngFile, cnf.CropCoordinates, se.tmpPngDir)
	if err != nil {
		return "", errors.Errorf(err, "图片裁剪失败")
	}

	var codeScanRes string
	switch cnf.CodeType {
	case common.CodeTypeQRCode:
		codeScanRes, err = util.QrCodeScan(croppedPngFile)
	case common.CodeTypeBarcode128:
		codeScanRes, err = util.Barcode128Scan(croppedPngFile)
	default:
		errors.Errorf(nil, "暂不支持的code类型:%s", cnf.CodeType)
	}
	return codeScanRes, nil
}

// 使用unipdf解析
func (se *SingleFileExtractor) extractWithRegByUnipdf(cnf *ExtractConfig) (string, error) {
	resource, ok := se.resource[cnf.PageNum]
	if !ok {
		return "", errors.Errorf(nil, "配置中的页码不存在,cnf.PageNum:%d, file:%s", cnf.PageNum, se.filePath)
	}
	text, err := util.NewUniPdf().ExtractText(resource.filePath, "", []int{})
	if err != nil {
		return "", errors.Errorf(err, "unipdf解析文字出错")
	}

	if se.extractor.withDebug {
		logger.LoggerSugar.Debugf("--->>> unipdfText:%s", text)
	}

	if cnf.compiledRegExp == nil {
		cnf.compiledRegExp = regexp.MustCompile(cnf.RegExp)
	}

	res := cnf.compiledRegExp.FindSubmatch([]byte(text))
	if len(res) == 2 {
		hintValue := util.StringPurify(string(res[1]))
		return hintValue, nil
	}
	return "", errors.Errorf(nil, "unipdf+正则匹配文字出错")
}

// 使用pdftotext解析
func (se *SingleFileExtractor) extractWithRegByPdfToText(cnf *ExtractConfig) (string, error) {
	resource, ok := se.resource[cnf.PageNum]
	if !ok {
		return "", errors.Errorf(nil, "配置中的页码不存在,cnf.PageNum:%d, file:%s", cnf.PageNum, se.filePath)
	}
	text, err := xpdf.PdfToText(resource.filePath, path.Join(se.tmpDir, path.Base(resource.filePath)))
	if err != nil {
		return "", errors.Errorf(err, "xpdf解析文字出错")
	}

	if se.extractor.withDebug {
		logger.LoggerSugar.Debugf("--->>> xpdfText:%s", text)
	}

	if cnf.compiledRegExp == nil {
		cnf.compiledRegExp = regexp.MustCompile(cnf.RegExp)
	}

	res := cnf.compiledRegExp.FindSubmatch([]byte(text))
	if len(res) == 2 {
		hintValue := util.StringPurify(string(res[1]))
		return hintValue, nil
	}
	return "", errors.Errorf(nil, "xpdf+正则匹配文字出错")
}

// 使用ocr解析
func (se *SingleFileExtractor) extractWithRegByOcr(cnf *ExtractConfig) (string, error) {
	resource, ok := se.resource[cnf.PageNum]
	if !ok {
		return "", errors.Errorf(nil, "配置中的页码不存在,cnf.PageNum:%d, file:%s", cnf.PageNum, se.filePath)
	}

	cleanFilename := strings.TrimSuffix(path.Base(se.filePath), path.Ext(se.filePath))
	cleanFilename = fmt.Sprintf("%x", md5.Sum([]byte(cleanFilename)))

	imagesBytes, err := xpdf.PdfToPngBytes(resource.filePath, se.tmpPngDir, cleanFilename)
	if err != nil {
		return "", errors.Errorf(err, "pdf转png bytes失败")
	}
	if len(imagesBytes) > 1 {
		return "", errors.Errorf(nil, "pdf转png bytes结果大于1")
	}

	client := gosseract.NewClient()
	defer client.Close()
	err = client.SetImageFromBytes(imagesBytes[0])
	if err != nil {
		return "", errors.Errorf(err, "ocr client.SetImageFromBytes failed")
	}

	ocrText, err := client.Text()
	if err != nil {
		return "", errors.Errorf(err, "读取图片中的文字失败")
	}
	if se.extractor.withDebug {
		logger.LoggerSugar.Debugf("--->>> ocrText:%s", ocrText)
	}

	if cnf.compiledRegExp == nil {
		cnf.compiledRegExp = regexp.MustCompile(cnf.RegExp)
	}

	res := cnf.compiledRegExp.FindSubmatch([]byte(ocrText))
	if len(res) == 2 {
		hintValue := util.StringPurify(string(res[1]))
		return hintValue, nil
	}
	return "", errors.Errorf(nil, "ocr+正则匹配文字出错")
}

func (se *SingleFileExtractor) extractWithRegV2(cnf *ExtractConfig) (string, error) {
	if cnf.TextExtractTools == nil {
		cnf.TextExtractTools = defaultTextExtractTools
	}

	var text string
	var err error
	for i := 0; i < len(cnf.TextExtractTools); i++ {
		tool := cnf.TextExtractTools[i]
		switch tool {
		case txtToolUnipdf:
			text, err = se.extractWithRegByUnipdf(cnf)
		case txtToolXpdf:
			text, err = se.extractWithRegByPdfToText(cnf)
		case txtToolOcr:
			text, err = se.extractWithRegByOcr(cnf)
		}
		if err != nil {
			continue
		}
		if text != "" {
			return text, nil
		}
	}

	return "", errors.Errorf(nil, "正则解析文字结果为空")
}
func (se *SingleFileExtractor) extractWithReg(cnf *ExtractConfig) (string, error) {
	resource, ok := se.resource[cnf.PageNum]
	if !ok {
		return "", errors.Errorf(nil, "配置中的页码不存在,cnf.PageNum:%d, file:%s", cnf.PageNum, se.filePath)
	}

	if resource.extractedText == "" {
		text, err := util.NewUniPdf().ExtractText(resource.filePath, "", []int{})
		if err != nil {
			logger.ErrorfWithEnv(se.extractor.withDebug, "unipdf解析文字失败, filename:%s,err:%s", resource.filePath, err)
			xpdfText, err := xpdf.PdfToText(resource.filePath, path.Join(se.tmpDir, path.Base(resource.filePath)))
			if err != nil {
				logger.ErrorfWithEnv(se.extractor.withDebug, "pdftotext解析文字失败, filename:%s,err:%s", resource.filePath, err)
			}
			resource.extractedText = xpdfText
		} else {
			resource.extractedText = text
		}
		logger.DebugfWithEnv(se.extractor.withDebug, "--->>> extractedText:%s", resource.extractedText)
	}

	if cnf.compiledRegExp == nil {
		cnf.compiledRegExp = regexp.MustCompile(cnf.RegExp)
	}

	res := cnf.compiledRegExp.FindSubmatch([]byte(resource.extractedText))
	if len(res) == 2 {
		hintValue := util.StringPurify(string(res[1]))
		return hintValue, nil
	}

	cleanFilename := strings.TrimSuffix(path.Base(se.filePath), path.Ext(se.filePath))
	cleanFilename = fmt.Sprintf("%x", md5.Sum([]byte(cleanFilename)))

	// 未匹配到，使用ocr
	if resource.ocrText == "" {
		imagesBytes, err := xpdf.PdfToPngBytes(resource.filePath, se.tmpPngDir, cleanFilename)
		if err != nil {
			return "", errors.Errorf(err, "pdf转png bytes失败")
		}
		if len(imagesBytes) > 1 {
			return "", errors.Errorf(nil, "pdf转png bytes结果大于1")
		}

		client := gosseract.NewClient()
		defer client.Close()
		err = client.SetImageFromBytes(imagesBytes[0])
		if err != nil {
			return "", errors.Errorf(err, "ocr client.SetImageFromBytes failed")
		}

		ocrText, err := client.Text()
		if err != nil {
			return "", errors.Errorf(err, "读取图片中的文字失败")
		}
		//fmt.Println("-->>", ocrText)
		resource.ocrText = ocrText

		if se.extractor.withDebug {
			logger.LoggerSugar.Debugf("--->>> ocrText:%s", ocrText)
		}
	}

	ocrRegRes := cnf.compiledRegExp.FindSubmatch([]byte(resource.ocrText))
	if len(ocrRegRes) == 2 {
		hintValue := util.StringPurify(string(ocrRegRes[1]))
		return hintValue, nil
	}

	return "", nil
}

func (se *SingleFileExtractor) extractWithTET(cnf *ExtractConfig) (string, error) {

	if cnf.TetCoordinates == nil {
		return "", errors.Errorf(nil, "tet 坐标配置项为空")
	}

	resource, ok := se.resource[cnf.PageNum]
	if !ok {
		return "", errors.Errorf(nil, "配置中的页码不存在,cnf.PageNum:%d, file:%s", cnf.PageNum, se.filePath)
	}

	f, err := ioutil.TempFile(common.CurrentDir, "*.txt")
	if err != nil {
		return "", nil
		// todo: 处理error
		//return nil, errors.Errorf(err, "create tmp file failed")
	}
	f.Close()
	//fmt.Println("临时文件需要清理:", f.Name())
	defer os.Remove(f.Name())

	text, err := util.ExtractTextByCoordinate(resource.filePath, strings.Join(cnf.TetCoordinates, " "), f.Name())
	if err != nil {
		return "", nil
		// todo: 处理error
	}

	return text, nil
}

// isNeedSplit 配置中如果有页数大于1的，则说明需要拆分pdf
func isNeedSplit(cnf []*ExtractConfig) bool {
	for i := 0; i < len(cnf); i++ {
		if cnf[i].PageNum > 1 {
			return true
		}
	}
	return false
}
