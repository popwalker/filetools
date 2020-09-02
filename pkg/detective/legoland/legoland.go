package legoland

import (
	"os"
	"path"
	"strings"

	"invtools/common"
	"invtools/pkg/util"

	"invtools/utils"

	"invtools/utils/validator"

	"invtools/utils/errors"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/viper"
)

const (
	DetectiveNickName = "legoland"
)

type DetectiveLeGoland struct {
	PdfFileDir    string `json:"pdf_file_dir" valid:"required"`
	ActivityName  string `json:"activity_name" valid:"required"`
	EffectiveDate string `json:"effective_date" valid:"required"`
}

func NewDetective(dir, act, valid string) *DetectiveLeGoland {
	return &DetectiveLeGoland{dir, act, valid}
}

func (d *DetectiveLeGoland) Detect() error {
	// 校验
	if err := d.Validate(); err != nil {
		return errors.Errorf(err, "参数校验失败")
	}

	// read dir & get files
	fis, err := util.ReadDirFiles(d.PdfFileDir, common.ExtPDF)
	if err != nil {
		return errors.Errorf(err, "读取目录失败")
	}

	// report to stdOut
	d.ReportBeforeProcess(len(fis))

	err = d.detect(fis)
	if err != nil {
		return errors.Errorf(err, "检测legoland的PDF发生错误")
	}

	return nil
}

func (d *DetectiveLeGoland) Validate() error {
	ok, err := validator.ValidateStruct(d)
	if err != nil {
		return errors.Errorf(err, "validator validate error")
	}

	if !ok {
		return errors.Errorf(nil, "validator validate failed")
	}

	if ok := utils.CheckDirIsExist(d.PdfFileDir); !ok {
		return errors.Errorf(nil, "目标文件夹不存在,dir:[%s]", d.PdfFileDir)
	}
	return nil
}

// ReportBeforeProcess 处理前report一次信息
func (d *DetectiveLeGoland) ReportBeforeProcess(n int) {
	util.Printf("一共有%d张PDF文件\n", n)
}

func (d *DetectiveLeGoland) detect(fis []string) error {
	count := len(fis)
	uiprogress.Start()
	bar := uiprogress.AddBar(count).AppendCompleted().PrependElapsed()
	// prepend the current step to the bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return "processing: " + path.Base(fis[b.Current()-1])
	})

	var unexpectedFiles []string
	var fileMd5Keys = make(map[string]string)
	var repeatedFiles []string
	for bar.Incr() {
		fi := fis[bar.Current()-1]
		md5key, err := utils.ComputeMd5String(fi)
		if err != nil {
			unexpectedFiles = append(unexpectedFiles, fi)
			continue
		}

		if _, ok := fileMd5Keys[md5key]; ok {
			repeatedFiles = append(repeatedFiles, path.Base(fi))
			continue
		} else {
			fileMd5Keys[md5key] = path.Base(fi)
		}

		err = d.detectSinglePdf(fi)
		if err != nil {
			unexpectedFiles = append(unexpectedFiles, path.Base(fi))
		}
	}

	repeated := len(repeatedFiles)
	detectFailed := len(unexpectedFiles)
	success := count - detectFailed - repeated
	successRate := float64(100 * success / count)
	util.Printf("本次检测一共扫描了%d张文件,重复文件:%d张,检测失败:%d张,符合预期的有%d张,达标率:%.02f%%\n", count, repeated, detectFailed, success, successRate)
	if repeated > 0 {
		util.Printf("以下是重复的文件:\n%v\n", strings.Join(repeatedFiles, "\n"))
	}

	if detectFailed > 0 {
		util.Printf("以下是检测失败的:\n%v\n", strings.Join(unexpectedFiles, "\n"))
	}

	// classify files
	if err := d.classifyFile(repeatedFiles, unexpectedFiles); err != nil {
		util.Printf("归类文件失败,err:%v", err)
	}

	return nil
}

func (d *DetectiveLeGoland) classifyFile(repeatedFiles, failedFiles []string) error {
	classify := viper.GetBool(common.LeGolandFlagClassify)
	if !classify {
		return nil
	}

	if len(repeatedFiles) > 0 {
		repeatedDir := path.Join(d.PdfFileDir, "repeated_files")
		if err := utils.CheckAndMkDir(repeatedDir); err != nil {
			return errors.Errorf(err, "create directory failed")
		}

		for _, v := range repeatedFiles {
			oldFullPath := path.Join(d.PdfFileDir, v)
			if err := os.Rename(oldFullPath, path.Join(repeatedDir, v)); err != nil {
				return errors.Errorf(err, "move file failed,filename:%s", v)
			}
		}
	}

	if len(failedFiles) > 0 {
		unexpectedDir := path.Join(d.PdfFileDir, "unexpected_files")
		if err := utils.CheckAndMkDir(unexpectedDir); err != nil {
			return errors.Errorf(err, "create directory failed")
		}

		for _, v := range failedFiles {
			oldFullPath := path.Join(d.PdfFileDir, v)
			newFullPath := path.Join(unexpectedDir, v)
			if err := os.Rename(oldFullPath, newFullPath); err != nil {
				return errors.Errorf(err, "move file failed,filename:%s", v)
			}
		}
	}
	util.Printf("classify files finished!\n")
	return nil
}

// detectSinglePdf 检测单个pdf
func (d *DetectiveLeGoland) detectSinglePdf(filePath string) error {
	pdfContent, err := util.NewUniPdf().ExtractText(filePath, "", []int{})
	if err != nil {
		return errors.Errorf(err, "ExtractText from legoland pdf file failed")
	}

	// detect activity name
	if !d.detectActivityName(pdfContent) {
		return errors.Errorf(nil, "detect actName from pdf content failed")
	}

	// detect effective date
	if !d.detectEffectiveDate(pdfContent) {
		return errors.Errorf(nil, "detect effective date from pdf content failed")
	}

	return nil
}

func (d *DetectiveLeGoland) detectActivityName(content string) bool {
	startFlag := "Ticket:"
	endFlag := "Customer"

	if !strings.Contains(content, startFlag) || !strings.Contains(content, endFlag) {
		return false
	}

	idx1 := strings.Index(content, startFlag)
	idx2 := strings.Index(content, endFlag)

	// get actName from pdf text and trim space
	actName := content[idx1+len(startFlag) : idx2]
	actName = strings.NewReplacer(" ", "").Replace(actName)

	// trim space
	expectedActName := strings.NewReplacer(" ", "").Replace(d.ActivityName)

	// compare
	if !strings.EqualFold(actName, expectedActName) {
		return false
	}

	return true
}

func (d *DetectiveLeGoland) detectEffectiveDate(content string) bool {
	startFlag := "Valid from:"
	endFlag := "Order No"

	if !strings.Contains(content, startFlag) || !strings.Contains(content, endFlag) {
		return false
	}

	idx1 := strings.Index(content, startFlag)
	idx2 := strings.Index(content, endFlag)

	// get effective date and trim space
	effectiveDate := content[idx1+len(startFlag) : idx2]
	effectiveDate = strings.NewReplacer(" ", "").Replace(effectiveDate)

	// trim space
	expectedEffectiveDate := strings.NewReplacer(" ", "").Replace(d.EffectiveDate)

	if !strings.EqualFold(effectiveDate, expectedEffectiveDate) {
		return false
	}

	return true
}
