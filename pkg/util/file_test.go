package util

import (
	"fmt"
	"testing"
)

func TestReadDirFilesV3(t *testing.T) {
	type args struct {
		dirname string
		ext     []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "TestReadDirFilesV3",
			args: args{
				dirname: "/Users/rick/Downloads/kay/origin_pdf/taipei101-single-1140",
				ext:     []string{".pdf", ".png"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadDirFilesV3(tt.args.dirname, tt.args.ext...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadDirFilesV3() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("len:", len(got))
			fmt.Println(got[:100])
		})
	}
}
