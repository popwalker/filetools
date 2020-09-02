package mupdf

import (
	"testing"
)

func Test_pdfClean(t *testing.T) {
	type args struct {
		infile  string
		outfile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test_pdfClean",
			args: args{
				infile:  "./broken.pdf",
				outfile: "./repaired.pdf",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := pdfClean(tt.args.infile, tt.args.outfile); (err != nil) != tt.wantErr {
				t.Errorf("pdfClean() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
