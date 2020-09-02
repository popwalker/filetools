package util

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"os"
	"strings"
	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
)

func TestUniPdf_ExtractText(t *testing.T) {
	type args struct {
		inputPath string
		password  string
		pages     []int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestUniPdf_ExtractText",
			args: args{
				inputPath: "../../testdata/disneyHK.pdf",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UniPdf{}
			got, err := u.ExtractText(tt.args.inputPath, tt.args.password, tt.args.pages)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("got:%s", got)
		})
	}
}

func TestUniPdf_ExtractTextWithPages(t *testing.T) {
	type args struct {
		inputPath string
		password  string
		pages     []int
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "TestUniPdf_ExtractTextWithPages",
			args: args{
				inputPath: "../../testdata/disneyHK.pdf",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UniPdf{}
			got, err := u.ExtractTextWithPages(tt.args.inputPath, tt.args.password, tt.args.pages)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTextWithPages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("got:%s", strings.Join(got, "###########"))
		})
	}
}

func TestUniPdf_SplitIntoFiles(t *testing.T) {
	type args struct {
		filePath  string
		outputDir string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "TestUniPdf_SplitIntoFiles",
			args: args{
				filePath:  "../testdata/hongkong_disneyland.pdf",
				outputDir: "/Users/rick/Downloads",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UniPdf{}
			got, err := u.SplitIntoFiles(tt.args.filePath, tt.args.outputDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitIntoFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("got:", got)
		})
	}
}

func TestUniPdf_SplitIntoBytes(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int][]byte
		wantErr bool
	}{
		{
			name:    "TestUniPdf_SplitIntoBytes",
			args:    args{filePath: "../testdata/hongkong_disneyland.pdf"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UniPdf{}
			got, err := u.SplitIntoBytes(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitIntoBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for k, v := range got {
				fmt.Println("k:", k, "v:", len(v))
				f, err := os.Create(fmt.Sprintf("/Users/rick/Downloads/%d.pdf", k))
				if err != nil {
					t.Fatal("创建文件失败")
				}

				_, err = f.Write(v)
				if err != nil {
					t.Fatal("write failed")
				}

				f.Close()
			}
		})
	}
}

func TestUniPdf_ExtractImagesIntoFiles(t *testing.T) {
	type args struct {
		filePath  string
		outputDir string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "TestUniPdf_ExtractImagesIntoFiles",
			args: args{
				filePath:  "../testdata/legoland.pdf",
				outputDir: "/Users/rick/Downloads",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UniPdf{}
			got, err := u.ExtractImagesIntoFiles(tt.args.filePath, tt.args.outputDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractImagesIntoFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("got:", got)
		})
	}
}

func TestUniPdf_ExtractImagesIntoJpegBytes(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    [][]byte
		wantErr bool
	}{
		{
			name:    "TestUniPdf_ExtractImagesIntoJpegBytes",
			args:    args{filePath: "/Users/rick/Downloads/tomchen/test/Untitled-part 74.pdf"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UniPdf{}
			got, err := u.ExtractImagesIntoJpegBytes(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractImagesIntoBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, v := range got {
				img, err := jpeg.Decode(bytes.NewBuffer(v))
				if err != nil {
					t.Fatal(err)
				}
				//var img image.Image
				//jpeg.Encode(bytes.NewBuffer(v), &img, nil)
				// prepare BinaryBitmap
				bmp, err := gozxing.NewBinaryBitmapFromImage(img)
				if err != nil {
					t.Fatal(err)
				}

				result, err := oned.NewCode128Reader().Decode(bmp, nil)
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println("result:",result.String())


				//f, err := os.Create(fmt.Sprintf("/Users/rick/Downloads/%d.jpg", k))
				//if err != nil {
				//	t.Fatal(err)
				//}
				//_, err = f.Write(v)
				//if err != nil {
				//	t.Fatal(err)
				//}
			}
		})
	}
}
