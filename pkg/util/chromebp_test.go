package util

import (
	"testing"
)

func TestChromedpPrintPdf(t *testing.T) {
	type args struct {
		url string
		to  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestChromedpPrintPdf",
			args: args{
				url: "https://www.puroland.jp/qrticket/e/?p=3000089253600221003006000000930000000000000008011100034056502000000640002020000000000000020190152797",
				to:  "./result.pdf",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ChromedpPrintPdf(tt.args.url, tt.args.to); (err != nil) != tt.wantErr {
				t.Errorf("ChromedpPrintPdf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
