package linktopdf

import (
	"fmt"
	"reflect"
	"testing"
)

func TestExecute(t *testing.T) {
	type args struct {
		input        string
		output       string
		concurrency  int
		needCompress bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestExecute",
			args: args{
				input:        "../../testdata/linktopdf.txt",
				output:       "./",
				concurrency:  2,
				needCompress: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Execute(tt.args.input, tt.args.output, tt.args.concurrency, tt.args.needCompress); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_divideLinksIntoGroup(t *testing.T) {
	type args struct {
		links    []string
		perGroup int
	}
	tests := []struct {
		name       string
		args       args
		wantGroups [][]string
	}{
		{
			name: "Test_divideLinksIntoGroup_1",
			args: args{
				links:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				perGroup: 1,
			},
			wantGroups: [][]string{{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}},
		},
		{
			name: "Test_divideLinksIntoGroup_2",
			args: args{
				links:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				perGroup: 2,
			},
			wantGroups: [][]string{{"1", "2"}, {"3", "4"}, {"5", "6"}, {"7", "8"}, {"9", "10"}},
		},
		{
			name: "Test_divideLinksIntoGroup_3",
			args: args{
				links:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				perGroup: 3,
			},
			wantGroups: [][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}, {"10"}},
		},
		{
			name: "Test_divideLinksIntoGroup_4",
			args: args{
				links:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				perGroup: 4,
			},
			wantGroups: [][]string{{"1", "2", "3", "4"}, {"5", "6", "7", "8"}, {"9", "10"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotGroups := divideLinksIntoGroup(tt.args.links, tt.args.perGroup); !reflect.DeepEqual(gotGroups, tt.wantGroups) {
				t.Errorf("divideLinksIntoGroup() = %v, want %v", gotGroups, tt.wantGroups)
			}
		})
	}
}

func Test_divideLinksIntoGroupV2(t *testing.T) {
	type args struct {
		links      []string
		groupCount int
	}
	tests := []struct {
		name       string
		args       args
		wantGroups [][]string
	}{
		{
			name: "Test_divideLinksIntoGroupV2_1",
			args: args{
				links:      []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				groupCount: 1,
			},
			wantGroups: [][]string{{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}},
		},
		{
			name: "Test_divideLinksIntoGroupV2_2",
			args: args{
				links:      []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				groupCount: 2,
			},
			wantGroups: [][]string{{"1", "2", "3", "4", "5"}, {"6", "7", "8", "9", "10"}},
		},
		{
			name: "Test_divideLinksIntoGroupV2_3",
			args: args{
				links:      []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				groupCount: 3,
			},
			wantGroups: [][]string{{"1", "2", "3", "4"}, {"5", "6", "7", "8"}, {"9", "10"}},
		},
		{
			name: "Test_divideLinksIntoGroupV2_4",
			args: args{
				links:      []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				groupCount: 4,
			},
			wantGroups: [][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}, {"10"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotGroups := divideLinksIntoGroupV2(tt.args.links, tt.args.groupCount); !reflect.DeepEqual(gotGroups, tt.wantGroups) {
				t.Errorf("divideLinksIntoGroupV2() = %v, want %v", gotGroups, tt.wantGroups)
			}
		})
	}
}

func Test_divideLinksIntoGroupV3(t *testing.T) {
	links := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	type args struct {
		links      []string
		groupCount int
	}
	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{
			name: "Test_divideLinksIntoGroupV3_1",
			args: args{
				links:      links,
				groupCount: 1,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV3_2",
			args: args{
				links:      links,
				groupCount: 2,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV3_3",
			args: args{
				links:      links,
				groupCount: 3,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV3_4",
			args: args{
				links:      links,
				groupCount: 4,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV3_5",
			args: args{
				links:      links,
				groupCount: 5,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV3_6",
			args: args{
				links:      links,
				groupCount: 6,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV3_7",
			args: args{
				links:      links,
				groupCount: 7,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV3_11",
			args: args{
				links:      links,
				groupCount: 11,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := divideLinksIntoGroupV3(tt.args.links, tt.args.groupCount)
			fmt.Println("-->>got:", got)
			if len(got) != tt.args.groupCount {
				t.Fatalf("group By failed")
			}
		})
	}
}

func Test_divideLinksIntoGroupV4(t *testing.T) {
	links := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	type args struct {
		links      []string
		groupCount int
	}
	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{
			name: "Test_divideLinksIntoGroupV4_1",
			args: args{
				links:      links,
				groupCount: 1,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV4_2",
			args: args{
				links:      links,
				groupCount: 2,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV4_3",
			args: args{
				links:      links,
				groupCount: 3,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV4_4",
			args: args{
				links:      links,
				groupCount: 4,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV4_5",
			args: args{
				links:      links,
				groupCount: 5,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV4_6",
			args: args{
				links:      links,
				groupCount: 6,
			},
		},
		{
			name: "Test_divideLinksIntoGroupV4_7",
			args: args{
				links:      links,
				groupCount: 7,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := divideLinksIntoGroupV4(tt.args.links, tt.args.groupCount)
			fmt.Println("-->>got:", got)
			if len(got) != tt.args.groupCount {
				t.Fatalf("group By failed")
			}
		})
	}
}

func Test_getOutput(t *testing.T) {
	type args struct {
		output string
	}
	tests := []struct {
		name     string
		args     args
		wantDir  string
		wantName string
		wantErr  bool
	}{
		{
			name:     "Test_getOutput_1",
			args:     args{output: "/Users/rick/a.zip"},
			wantDir:  "/Users/rick",
			wantName: "a.zip",
			wantErr:  false,
		},
		{
			name:     "Test_getOutput_2",
			args:     args{output: "/Users/rick"},
			wantDir:  "/Users/rick",
			wantName: "xxx.zip",
			wantErr:  false,
		},
		{
			name:     "Test_getOutput_3",
			args:     args{output: "rick/b.zip"},
			wantDir:  "/Users/rick",
			wantName: "b.zip",
			wantErr:  false,
		},
		{
			name:     "Test_getOutput_4",
			args:     args{output: "rick/"},
			wantDir:  "/Users/rick",
			wantName: "xxx.zip",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDir, gotName, err := getOutput(tt.args.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDir != tt.wantDir {
				t.Errorf("getOutput() gotDir = %v, want %v", gotDir, tt.wantDir)
			}
			if gotName != tt.wantName {
				t.Errorf("getOutput() gotName = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func Test_genFileNameFromURL(t *testing.T) {
	u := "https://www.puroland.jp/qrticket/e/?p=3000089253600423003004000000930000000000000008011100034056501000010640002020000000000000020190152798"
	type args struct {
		u string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Test_genFileNameFromURL",
			args:    args{u},
			wantErr: false,
		},
		{
			name:    "Test_genFileNameFromURL",
			args:    args{"https://www.baidu.com"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := genFileNameFromURL(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("genFileNameFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("got:%v", got)
		})
	}
}
