package legoland

import (
	"testing"

	"invtools/pkg/util"
)

func TestDetectiveLeGoland_detectActivityName(t *testing.T) {
	content, err := util.NewUniPdf().ExtractText("../../../testdata/legoland.pdf", "", []int{})
	if err != nil {
		t.Fatalf("extract text content from legoland pdf failed. err:%v", err)
	}
	type fields struct {
		PdfFileDir    string
		ActivityName  string
		EffectiveDate string
	}
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "TestDetectiveLeGoland_detectActivityName",
			fields: fields{
				ActivityName: "COMBO TRD 2D TP + WP + SLC (C/S) OPEN",
			},
			args: args{content: content},
			want: true,
		},
		{
			name: "TestDetectiveLeGoland_detectActivityName",
			fields: fields{
				ActivityName: "TEST_WRONG_COMBO TRD 2D TP + WP + SLC (C/S) OPEN",
			},
			args: args{content: content},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DetectiveLeGoland{
				PdfFileDir:    tt.fields.PdfFileDir,
				ActivityName:  tt.fields.ActivityName,
				EffectiveDate: tt.fields.EffectiveDate,
			}
			got := d.detectActivityName(tt.args.content)
			if got != tt.want {
				t.Errorf("detectActivityName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectiveLeGoland_detectEffectiveDate(t *testing.T) {
	content, err := util.NewUniPdf().ExtractText("../../../testdata/legoland.pdf", "", []int{})
	if err != nil {
		t.Fatalf("extract text content from legoland pdf failed. err:%v", err)
	}

	type fields struct {
		PdfFileDir    string
		ActivityName  string
		EffectiveDate string
	}
	type args struct {
		content string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "TestDetectiveLeGoland_detectEffectiveDate",
			fields: fields{
				EffectiveDate: "25/07/2019 thru 25/01/2020",
			},
			args: args{content: content},
			want: true,
		},
		{
			name: "TestDetectiveLeGoland_detectEffectiveDate",
			fields: fields{
				EffectiveDate: "TEST_WRONG_25/07/2019 thru 25/01/2020",
			},
			args: args{content: content},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DetectiveLeGoland{
				PdfFileDir:    tt.fields.PdfFileDir,
				ActivityName:  tt.fields.ActivityName,
				EffectiveDate: tt.fields.EffectiveDate,
			}
			if got := d.detectEffectiveDate(tt.args.content); got != tt.want {
				t.Errorf("detectEffectiveDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
