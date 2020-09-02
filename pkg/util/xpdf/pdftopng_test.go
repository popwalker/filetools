package xpdf

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPdfToPng(t *testing.T) {
	type args struct {
		pdfFilePath string
		outputDir   string
		prefix      string
	}
	tests := []struct {
		name    string
		args    args
		want    [][]byte
		wantErr bool
	}{
		{
			name: "TestPdfToPng",
			args: args{
				pdfFilePath: "/Users/rick/Downloads/KLK_dev_correct.pdf",
				outputDir:   "/Users/rick/Downloads/KLK_dev_correct",
				prefix:      "aaa",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PdfToPng(tt.args.pdfFilePath, tt.args.outputDir, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("PdfToPng() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PdfToPng() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPdfToPngV2(t *testing.T) {
	type args struct {
		pdfFilePath string
		outputDir   string
		prefix      string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "TestPdfToPngV2",
			args: args{
				pdfFilePath: "../testdata/disneyHK.pdf",
				outputDir:   "/Users/rick/Downloads",
				prefix:      "ababababababab",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PdfToPngV2(tt.args.pdfFilePath, tt.args.outputDir, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("PdfToPngV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
		})
	}
}
