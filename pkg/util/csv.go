package util

import (
	"encoding/csv"
	"os"
	"strings"

	"invtools/common"

	"invtools/utils/errors"

	"github.com/tealeg/xlsx"
)

type TableWriter struct {
	file      string
	csvWriter *csv.Writer
	xlsxSheet *xlsx.Sheet
	xlsxFile  *xlsx.File
}

func NewTableWriter(input string) *TableWriter {
	return &TableWriter{
		file: input,
	}
}

func (w *TableWriter) DecideWriter() error {

	if strings.HasSuffix(w.file, common.ExtCsv) {
		f, err := os.OpenFile(w.file, os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			return errors.Errorf(err, "创建output文件失败")
		}
		w.csvWriter = csv.NewWriter(f)
	} else if strings.HasSuffix(w.file, common.ExtExecl) {
		f := xlsx.NewFile()
		//if err != nil {
		//    return errors.Errorf(err, "xlsx open file failed")
		//}
		w.xlsxFile = f
		sheet, err := f.AddSheet("Sheet1")
		if err != nil {
			return errors.Errorf(err, "xlsx add sheet failed")
		}
		w.xlsxSheet = sheet
	}

	return nil
}

func (w *TableWriter) WriteRecord(record []string) error {
	switch {
	case w.csvWriter != nil:
		return WriteRecordToCsv(w.csvWriter, record)
	case w.xlsxSheet != nil:
		return WriteRecordToXlsx(w.xlsxSheet, record)
	}

	return nil
}

func (w *TableWriter) Close() error {
	switch {
	case w.csvWriter != nil:
		if err := w.csvWriter.Error(); err != nil {
			return errors.Errorf(err, "csv writer failed")
		}
		w.csvWriter.Flush()
	case w.xlsxSheet != nil:
		if err := w.xlsxFile.Save(w.file); err != nil {
			return errors.Errorf(err, "xlsx save failed")
		}
	}
	return nil
}

func WriteRecordToCsv(w *csv.Writer, record []string) error {
	if err := w.Write(record); err != nil {
		return errors.Errorf(err, "write record to csv failed")
	}

	return nil
}

func WriteRecordToXlsx(sheet *xlsx.Sheet, record []string) error {
	row := sheet.AddRow()
	for _, v := range record {
		cell := row.AddCell()
		cell.Value = v
	}

	return nil
}
