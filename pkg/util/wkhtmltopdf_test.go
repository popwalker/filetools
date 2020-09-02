package util

import (
	"testing"
)

func TestWkHtmlToPDf(t *testing.T) {
	type args struct {
		reqURL  string
		pdfFile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestWkHtmlToPDf",
			args: args{
				reqURL:  "https://www.baidu.com",
				pdfFile: "./result.pdf",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WkHtmlToPDf(tt.args.reqURL, tt.args.pdfFile); (err != nil) != tt.wantErr {
				t.Errorf("WkHtmlToPDf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

	}
}
