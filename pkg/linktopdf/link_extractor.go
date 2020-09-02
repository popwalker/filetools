package linktopdf

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"invtools/common"
	"invtools/pkg/util"

	"invtools/utils"

	"invtools/utils/errors"

	"github.com/tealeg/xlsx"
)

// LinkExtractor extracts links from input file
type LinkExtractor interface {
	Extract() ([]string, error)
}

func NewLinkExtractor(input string) (LinkExtractor, error) {
	if input == "" {
		return nil, errors.Errorf(nil, "input file is empty")
	}

	var e LinkExtractor

	ext := path.Ext(input)
	switch ext {
	case common.ExtCsv:
		e = &CsvExtractor{input}
	case common.ExtExecl:
		e = &ExcelExtractor{input}
	case common.ExtTxt:
		e = &TxtExtractor{input}
	default:
		return nil, errors.Errorf(nil, "unsupported file extension:%s", ext)
	}

	return e, nil
}

type CsvExtractor struct {
	Input string
}

func (e *CsvExtractor) Extract() (links []string, err error) {
	if ok := utils.CheckFileIsExist(e.Input); !ok {
		err = errors.Errorf(nil, "input file does not exists!")
		return
	}

	f, err := os.OpenFile(e.Input, os.O_RDONLY, 0644)
	if err != nil {
		err = errors.Errorf(err, "cannot open input file,err:%v", err)
		return
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		err = errors.Errorf(err, "csv readAll failed")
		return
	}

	for _, record := range records {
		if len(record) > 0 {
			if record[0] != "" && strings.Contains(record[0], "http") {
				links = append(links, record[0])
			}
		}
	}

	repeated := util.CheckRepeat(links)
	if len(repeated) != 0 {
		fmt.Printf("[linktopdf] 发现有重复的链接，请检查. 以下是重复链接\n%s\n", strings.Join(repeated, "\n"))
	}

	return
}

type ExcelExtractor struct {
	Input string
}

func (e *ExcelExtractor) Extract() (links []string, err error) {
	if ok := utils.CheckFileIsExist(e.Input); !ok {
		err = errors.Errorf(nil, "input file does not exists!")
		return
	}

	// open excel
	f, err := xlsx.OpenFile(e.Input)
	if err != nil {
		err = errors.Errorf(err, "xlsx open file failed")
		return
	}

	if len(f.Sheets) == 0 {
		err = errors.Errorf(nil, "excel file has no sheet,please check!")
		return
	}

	sheet := f.Sheets[0]
	for _, row := range sheet.Rows {
		if len(row.Cells) > 0 {
			link := row.Cells[0].String()
			if link != "" && strings.Contains(link, "http") {
				links = append(links, link)
			}
		}
	}

	repeated := util.CheckRepeat(links)
	if len(repeated) != 0 {
		fmt.Printf("[linktopdf] 发现有重复的链接，请检查. 以下是重复链接\n%s\n", strings.Join(repeated, "\n"))
	}

	return
}

type TxtExtractor struct {
	Input string
}

func (e *TxtExtractor) Extract() (links []string, err error) {
	if ok := utils.CheckFileIsExist(e.Input); !ok {
		err = errors.Errorf(nil, "input file does not exists!")
		return
	}

	f, err := os.OpenFile(e.Input, os.O_RDONLY, 0644)
	if err != nil {
		err = errors.Errorf(err, "open txt file failed")
		return
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		err = errors.Errorf(err, "read txt file failed")
		return
	}

	rows := strings.Split(string(b), "\n")
	for _, row := range rows {
		if row != "" && strings.Contains(row, "http") {
			links = append(links, row)
		}
	}

	repeated := util.CheckRepeat(links)
	if len(repeated) != 0 {
		fmt.Printf("[linktopdf] 发现有重复的链接，请检查. 以下是重复链接\n%s\n", strings.Join(repeated, "\n"))
	}

	return
}
