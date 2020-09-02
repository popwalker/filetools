package xpdf

import (
	"fmt"
	"testing"
)

func TestPdfToText(t *testing.T) {
	type args struct {
		pdfFilePath    string
		outputFilePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestPdfToText",
			args: args{
				pdfFilePath:    "/Users/rick/Downloads/sonia/4100_5713/20518442_23774158_e31e195c52d54aa46565ce6d22899e36_1.pdf",
				outputFilePath: "/Users/rick/Downloads/sonia/pdftotext.txt",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PdfToText(tt.args.pdfFilePath, tt.args.outputFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("PdfToText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("----> text:", got)
		})
	}
}
