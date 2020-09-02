package compatible

import (
	"testing"
)

func Test_getOutputFileFromRcv(t *testing.T) {
	type args struct {
		originInputFilename  string
		originOutputFilename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name:"Test_getOutputFileFromRcv",
			args:args{
				originInputFilename:  "/Users/rick/Downloads/lester1/pdf/RCV20200302005138929",
				originOutputFilename: "/Users/rick/Downloads/lester1/csv/abcd.xlsx",
			},
			want:"/Users/rick/Downloads/lester1/csv/abcd.xlsx",
		},
		{
			name:"Test_getOutputFileFromRcv",
			args:args{
				originInputFilename:  "/Users/rick/Downloads/lester1/pdf/RCV20200302005138929",
				originOutputFilename: "/Users/rick/Downloads/lester1/csv/system.xlsx",
			},
			want:"/Users/rick/Downloads/lester1/csv/RCV20200302005138929.xlsx",
		},
		{
			name:"Test_getOutputFileFromRcv",
			args:args{
				originInputFilename:  "/Users/rick/Downloads/lester1/pdf/RCV20200302005138929",
				originOutputFilename: "/Users/rick/Downloads/lester1/csv/system.xlsx",
			},
			want:"/Users/rick/Downloads/lester1/csv/RCV20200302005138929.xlsx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOutputFileFromRcv(tt.args.originInputFilename, tt.args.originOutputFilename); got != tt.want {
				t.Errorf("getOutputFileFromRcv() = %v, want %v", got, tt.want)
			}
		})
	}
}
