package linktopdf

import (
	"fmt"
	"testing"
)

func TestCsvExtractor_Extract(t *testing.T) {
	e, err := NewLinkExtractor("../../testdata/linktopdf.csv")
	if err != nil {
		t.Fatalf("NewLinkExtractor failed, err:%v", err)
	}

	tests := []struct {
		name      string
		wantLinks []string
		wantErr   bool
	}{
		{
			name:    "TestCsvExtractor_Extract",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, err := e.Extract()
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			fmt.Printf("got %d links:%v", len(gotLinks), gotLinks)
		})
	}
}

func TestExcelExtractor_Extract(t *testing.T) {
	e, err := NewLinkExtractor("../../testdata/linktopdf.xlsx")
	if err != nil {
		t.Fatalf("NewLinkExtractor failed, err:%v", err)
	}

	tests := []struct {
		name      string
		wantLinks []string
		wantErr   bool
	}{
		{
			name:    "TestExcelExtractor_Extract",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, err := e.Extract()
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			fmt.Printf("got %d links:%v", len(gotLinks), gotLinks)
		})
	}
}

func TestTxtExtractor_Extract(t *testing.T) {
	e, err := NewLinkExtractor("../../testdata/linktopdf.txt")
	if err != nil {
		t.Fatalf("NewLinkExtractor failed, err:%v", err)
	}

	tests := []struct {
		name      string
		wantLinks []string
		wantErr   bool
	}{
		{
			name:    "TestTxtExtractor_Extract",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, err := e.Extract()
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			fmt.Printf("got %d links:%v", len(gotLinks), gotLinks)
		})
	}
}
