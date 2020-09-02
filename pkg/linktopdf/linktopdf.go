package linktopdf

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"invtools/common"
	"invtools/pkg/util"

	"invtools/utils"

	"invtools/utils/errors"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/viper"
)

const (
	LinktopdfName = "Linktopdf"
)

func Execute(input, output string, concurrency int, needCompress bool) error {
	if ok := utils.CheckFileIsExist(input); !ok {
		return errors.Errorf(nil, "input file not exists")
	}

	e, err := NewLinkExtractor(input)
	if err != nil {
		return errors.Errorf(err, "new NewLinkExtractor failed")
	}

	links, err := e.Extract()
	if err != nil || len(links) == 0 {
		return errors.Errorf(err, "extract links from input file failed")
	}

	dir, zipFileName, err := getOutput(output)
	if err != nil {
		return errors.Errorf(err, "detect output failed")
	}

	count := len(links)
	fmt.Printf("[linktopdf] 检测到%d个链接，即将开始打印\n", count)

	batchGrp := divideLinksIntoGroupV3(links, concurrency)

	uiprogress.Start()
	var (
		wg sync.WaitGroup
		//waitTime       = time.Millisecond * 100
		successPrinted []string
		failedPrinted  []string
		st             = time.Now()
	)

	wg.Add(len(batchGrp))
	for _, value := range batchGrp {
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
				u := grp[bar.Current()-1]
				filePath, err := printPdf(u, dir)
				if err != nil {
					failedPrinted = append(failedPrinted, fmt.Sprintf("URL:%s, err:%v", u, err))
				} else {
					successPrinted = append(successPrinted, filePath)
				}
			}
		})(&grp)
	}

	wg.Wait()
	uiprogress.Stop()

	fmt.Printf("[linktopdf] 本次一共生成%d个pdf，失败%d个,总耗时:%s\n", len(successPrinted), len(failedPrinted), time.Since(st))
	if len(failedPrinted) > 0 {
		fmt.Println("[linktopdf] 发生错误:", strings.Join(failedPrinted, "\n"))
	}

	fmt.Printf("[linktopdf] pdf保存目录:%s\n", dir)

	if !needCompress {
		return nil
	}

	st = time.Now()
	fmt.Println("[linktopdf] 打印pdf完毕，准备压缩!")
	err = compress(dir, path.Join(path.Dir(dir), zipFileName))
	if err != nil {
		return errors.Errorf(err, "压缩失败")
	}

	fmt.Printf("[linktopdf] 打包压缩完毕! 耗时:%v, 压缩文件:%s\n", time.Since(st), zipFileName)

	if len(failedPrinted) > 0 {
		fmt.Printf("以下打印失败的链接:\n%s\n", strings.Join(failedPrinted, "\n"))
	}

	return nil
}

// compress 打包压缩
func compress(fileDir, zipFilePath string) error {
	var (
		buf       = new(bytes.Buffer)
		w         = zip.NewWriter(buf)
		hm        = make(map[string]struct{})
		repeatedF []string
	)

	// read dir
	fis, err := ioutil.ReadDir(fileDir)
	if err != nil {
		return errors.Errorf(err, "readDir failed,dir:%s", fileDir)
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		// ignore unexpected file extension
		if strings.ToLower(path.Ext(fi.Name())) != common.ExtPDF {
			continue
		}

		// create buf writer
		wf, err := w.Create(fi.Name())
		if err != nil {
			return errors.Errorf(err, "zip writer create failed,fileName:[%s]", fi.Name())
		}

		// open file
		f, err := os.OpenFile(path.Join(fileDir, fi.Name()), os.O_RDONLY, 0644)
		if err != nil {
			return errors.Errorf(err, "open pdf file failed")
		}

		// write to buf writer
		_, err = io.Copy(wf, f)
		if err != nil {
			return errors.Errorf(err, "write file into zip file failed,file:[%s]", fi.Name())
		}

		// get md5 of this file
		hs, err := computeMd5(f)
		if err != nil {
			return errors.Errorf(err, "compute md5 failed")
		}

		// detect repeat files
		if _, ok := hm[hs]; ok {
			repeatedF = append(repeatedF, f.Name())
		} else {
			hm[hs] = struct{}{}
		}

		// close file
		if err := f.Close(); err != nil {
			return errors.Errorf(err, "关闭pdf文件失败")
		}
	}

	// close buf writer
	err = w.Close()
	if err != nil {
		return errors.Errorf(err, "close zip file writer failed")
	}

	// create a zip file
	zipf, err := os.OpenFile(zipFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC|os.O_APPEND, 0644)
	if err != nil {
		return errors.Errorf(err, "create zip file failed")
	}
	defer zipf.Close()

	// copy write buf to zip file
	_, err = io.Copy(zipf, buf)
	if err != nil {
		return errors.Errorf(err, "write buf into zip file failed")
	}

	if len(repeatedF) > 0 {
		fmt.Printf("[linktopdf] 压缩时校验重复，发现以下重复文件:\n%s\n", strings.Join(repeatedF, "\n"))
	}

	// delete files
	err = utils.RmAll(fileDir)
	if err != nil {
		return errors.Errorf(err, "删除pdf文件失败")
	}

	return nil
}

func computeMd5(f *os.File) (string, error) {
	if f == nil {
		return "", errors.Errorf(nil, "param f is nil")
	}

	// reset file offset
	if ret, err := f.Seek(0, 0); err != nil {
		return "", errors.Errorf(err, "after write into buf writer,reset file offset failed.ret:%d", ret)
	}

	// get hash value to detect repeated file
	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", errors.Errorf(err, "copy file to hash failed")
	}

	var b []byte
	hvs := util.UnsafeByte2str(hash.Sum(b))

	return hvs, nil
}

func printPdf(u, dir string) (string, error) {

	fileName, err := genFileNameFromURL(u)
	if err != nil {
		return "", errors.Errorf(err, "根据url生成文件名失败")
	}

	filepath := path.Join(dir, fileName)

	// pdf资源直接下载
	if strings.Contains(u, ".pdf") || strings.Contains(u, "skybus.umd.com.au") {
		if err := downloadPdf(u, filepath); err != nil {
			return "", errors.Errorf(err, "下载pdf文件失败")
		}
		return filepath, nil
	}

	// 决定打印形式(html转pdf)
	printType := viper.GetString(common.LinkToPdfFlagPrintType)
	switch printType {
	case common.PrintTypeChromedp:
		err = util.ChromedpPrintPdf(u, filepath)
	case common.PrintTypeWkhtmltopdf:
		err = util.WkHtmlToPDf(u, filepath)
	default:
		err = errors.Errorf(nil, "unexpected print type:%s", printType)
	}

	if err != nil {
		return "", errors.Errorf(err, "根据url打印pdf失败")
	}
	return filepath, nil
}

func downloadPdf(u string, filePath string) error {
	if !strings.Contains(u, "http") {
		return errors.Errorf(nil, "url不合法,url:[%s]", u)
	}

	resp, err := http.Get(u)
	if err != nil {
		return errors.Errorf(err, "http请求URL失败,url:%s", u)

	}
	defer resp.Body.Close()

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return errors.Errorf(err, "打开文件失败,filePath:%s", filePath)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return errors.Errorf(err, "写入数据流到文件失败")
	}

	return nil
}

func genFileNameFromURL(u string) (string, error) {
	var (
		fileName string
		err      error
	)

	URL, err := url.Parse(u)
	if err != nil {
		return "", errors.Errorf(err, "parse url failed")
	}

	switch {
	case strings.Contains(u, "www.puroland.jp"):
		// puroland的url直接取参数p
		fileName = URL.Query().Get("p")
	case strings.Contains(u, "skybus.umd.com.au"):
		// 根据url:https://skybus.umd.com.au/skybus/bulk/template/generic/12170/1234567890/, 组合文件名:skybus_1234567890.pdf
		fileName = fmt.Sprintf("skybus_%s", path.Base(URL.Path))
	case strings.Contains(u, ".pdf"):
		// 根据url:https://www.xxx.com/upload_voucher/2019/08/01/xxx.pdf, 得到文件名: xxx
		fileName = strings.Replace(GetFileNameFromPdfURL(u), ".pdf", "", -1)
	case strings.Contains(u, "voucher/KLK"):
		// 根据url:https://www.xxx.com/zh-CN/voucher/KLK0877301055?token=5005f4db-8cde-4f3b-74b4-09493e6b3739&lang=zh_CN,得到文件名
		uRL, _ := url.Parse(u)
		token := uRL.Query().Get("token")
		fileName = token
	default:
		fileName = util.ComputeMd5String(u)
	}

	// 添加.pdf文件后缀
	return fmt.Sprintf("%s.pdf", fileName), nil
}

func GetFileNameFromPdfURL(u string) string {
	return path.Base(u)
}

/*
/path/to/file.zip
/path/to/dir
file.zip
to/dir/file.zip
*/
func getOutput(output string) (dir, name string, err error) {
	now := time.Now().In(utils.LocationCST)
	ok := path.IsAbs(output)

	if ok {
		ext := path.Ext(output)
		if ext != "" && strings.EqualFold(common.ExtZip, ext) {
			dir = path.Dir(output)
			name = path.Base(output)
		} else {
			dir = output
			name = getRandZipFileName(now)
		}
	} else {
		absOutput, err := filepath.Abs(output)
		if err != nil {
			return "", "", errors.Errorf(err, "尝试将相对路径转换为绝对路径失败")
		}

		dir = common.CurrentDir
		ext := path.Ext(absOutput)
		if ext != "" && strings.EqualFold(common.ExtZip, ext) {
			name = path.Base(output)
		} else {
			dir = absOutput
			name = getRandZipFileName(now)
		}
	}

	dir = path.Join(dir, fmt.Sprintf("linktopdf_%s", now.Format("20060102150405")))

	err = utils.CheckAndMkDir(dir)
	if err != nil {
		err = errors.Errorf(err, "mkdir failed")
		return
	}

	return
}

func getRandZipFileName(t time.Time) string {
	return fmt.Sprintf("linktopdf_%s.zip", t.Format("20060102150405"))
}

func divideLinksIntoGroup(links []string, perGroup int) (groups [][]string) {
	count := len(links)
	remainder := count % perGroup

	if perGroup == 1 {
		groups = append(groups, links)
		return
	}

	// 按照并发量分组
	for i := 0; i < count; i += perGroup {
		step := i + perGroup
		var linksGrp []string
		if remainder != 0 && count-i == remainder {
			linksGrp = links[i:]
		} else {
			linksGrp = links[i:step]
		}

		groups = append(groups, linksGrp)
	}
	return
}

func divideLinksIntoGroupV2(links []string, groupCount int) (groups [][]string) {
	count := len(links)
	perGroup := int(math.Ceil(float64(count) / float64(groupCount)))
	remainder := count % groupCount
	if remainder != 0 {
		groupCount += 1
	}

	if groupCount == 1 {
		groups = append(groups, links)
		return
	}

	for i := 0; i < count; i += perGroup {
		step := i + perGroup
		var linksGrp []string

		if remainder != 0 && step > count {
			linksGrp = links[i:]
		} else {
			linksGrp = links[i:step]
		}

		groups = append(groups, linksGrp)
	}

	return
}

// divideLinksIntoGroupV3 将一个数组中的元素，均匀的分散到N组中
func divideLinksIntoGroupV3(links []string, groupCount int) [][]string {
	count := len(links)
	if groupCount > count {
		groupCount = count
	}
	groups := make([][]string, groupCount)
	remainder := count % groupCount

	for i := 0; i < count; i += groupCount {
		step := i + groupCount
		var linksGrp []string
		if remainder != 0 && step > count {
			linksGrp = links[i:]
		} else {
			linksGrp = links[i:step]
		}

		for i, link := range linksGrp {
			groups[i] = append(groups[i], []string{link}...)
		}
	}

	return groups
}

func divideLinksIntoGroupV4(links []string, groupCount int) (groups [][]string) {
	count := len(links)
	remainder := count % groupCount
	perGroup := int(math.Floor(float64(count) / float64(groupCount)))

	for i := 0; i < count; i += perGroup {
		step := i + perGroup
		var linksGrp []string
		if remainder != 0 && step > count {
			linksGrp = links[i:]
			groups[len(groups)-1] = append(groups[len(groups)-1], linksGrp...)
		} else {
			linksGrp = links[i:step]
			if len(groups) == groupCount {
				groups[len(groups)-1] = append(groups[len(groups)-1], linksGrp...)
			} else {
				groups = append(groups, linksGrp)
			}
		}
	}

	return groups
}
