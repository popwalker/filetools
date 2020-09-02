package compatible

import (
	"sync"
	"testing"
)

func TestExtractor_extractWithOcr(t *testing.T) {
	type fields struct {
		inputDir       string
		outputFile     string
		rawArgs        []string
		concurrency    int
		maxReadPage    int
		withCoordinate bool
		withOcr        bool
		coordinates    sync.Map
		regexps        sync.Map
		resultKeys     []string
	}
	type args struct {
		result   *Result
		filePath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:"TestExtractor_extractWithOcr",
			fields:fields{
				inputDir:       "",
				outputFile:     "",
				rawArgs:        nil,
				concurrency:    0,
				maxReadPage:    0,
				withCoordinate: false,
				withOcr:        false,
				coordinates:    sync.Map{},
				regexps:        sync.Map{},
				resultKeys:     nil,
			},
			args:args{
				result:   nil,
				filePath: "",
			},
			wantErr:false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Extractor{
				inputDir:       tt.fields.inputDir,
				outputFile:     tt.fields.outputFile,
				rawArgs:        tt.fields.rawArgs,
				concurrency:    tt.fields.concurrency,
				maxReadPage:    tt.fields.maxReadPage,
				withCoordinate: tt.fields.withCoordinate,
				withOcr:        tt.fields.withOcr,
				coordinates:    tt.fields.coordinates,
				regexps:        tt.fields.regexps,
				resultKeys:     tt.fields.resultKeys,
			}
			if err := e.extractWithOcr(tt.args.result, tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("extractWithOcr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
