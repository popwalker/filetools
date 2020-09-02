package util

import (
	"encoding/csv"
	"testing"

	"github.com/tealeg/xlsx"
)

func TestTableWriter_WriteRecord(t *testing.T) {
	type fields struct {
		file       string
		csvWriter  *csv.Writer
		xlsxWriter *xlsx.Sheet
	}
	type args struct {
		record []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestTableWriter_WriteRecord_csv",
			fields: fields{
				file: "./a.csv",
			},
			args:    args{record: []string{"a", "b", "c"}},
			wantErr: false,
		},
		{
			name: "TestTableWriter_WriteRecord_xlsx",
			fields: fields{
				file: "./a.xlsx",
			},
			args:    args{record: []string{"d", "e", "f"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewTableWriter(tt.fields.file)
			defer w.Close()
			if err := w.DecideWriter(); err != nil {
				t.Fatalf("decide writer failed,err:%v", err)
			}

			if err := w.WriteRecord(tt.args.record); (err != nil) != tt.wantErr {
				t.Errorf("WriteRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
