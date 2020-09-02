package pdfcpu

import (
	"fmt"
	"testing"
)

func TestPdfToPng(t *testing.T) {
	type args struct {
		pdfFilePath string
		outputDir   string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "TestPdfToPng",
			args: args{
				pdfFilePath: "../../../testdata/Village_Roadshow_Theme_Parks.pdf",
				outputDir:   "../../../testdata/testoutput",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PdfToPng(tt.args.pdfFilePath, tt.args.outputDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("PdfToPng() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
		})
	}
}
