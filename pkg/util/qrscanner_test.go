package util

import (
	"testing"
)

func TestQrCodeScan(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name     string
		args     args
		wantCode string
		wantErr  bool
	}{
		{
			name:     "TestQrCodeScan",
			args:     args{input: "../../testdata/qrcode/qrcodepic-0c94f75.png"},
			wantCode: "qrcodepic-0c94f75",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, err := QrCodeScan(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("QrCodeScan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCode != tt.wantCode {
				t.Errorf("QrCodeScan() gotCode = %v, want %v", gotCode, tt.wantCode)
			}
		})
	}
}

func TestBarcode128Scan(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name     string
		args     args
		wantCode string
		wantErr  bool
	}{
		{
			name:     "TestBarcode128Scan",
			args:     args{input: "../../testdata/barcode128/I75_1_8.png"},
			wantCode: "barcode128pic-7cf34d2",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, err := Barcode128Scan(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Barcode128Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCode != tt.wantCode {
				t.Errorf("Barcode128Scan() gotCode = %v, want %v", gotCode, tt.wantCode)
			}
		})
	}
}
