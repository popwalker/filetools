package util

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"invtools/common"
)

func TestExtractTextByCoordinate(t *testing.T) {
	f, err := ioutil.TempFile(common.CurrentDir, "*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	type args struct {
		input      string
		coordinate string
		output     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestExtractTextByCoordinate",
			args: args{
				input:      "../testdata/legoland.pdf",
				coordinate: "38 707.93 243.91 716.93",
				output:     f.Name(),
			},
			want:    "Ticket: COMBO TRD 2D TP + WP + SLC (C/S) OPEN",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.Remove(tt.args.output)
			got, err := ExtractTextByCoordinate(tt.args.input, tt.args.coordinate, tt.args.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTextByCoordinate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Logf("got len:%d, want_len:%d", len(strings.TrimSpace(got)), len(tt.want))
				t.Errorf("ExtractTextByCoordinate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
