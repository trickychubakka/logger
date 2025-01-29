package main

import (
	"testing"
)

//func Test_createAnalysisTypesRegistry1(t *testing.T) {
//	var wantErr bool = false
//
//	_, err := createAnalysisTypesRegistry()
//	if (err != nil) != wantErr {
//		t.Errorf("readConfig() error = %v, wantErr %v", err, wantErr)
//		return
//	}
//}

func Test_createAnalysisTypesRegistry(t *testing.T) {
	tests := []struct {
		name string
		//want    analysisTypeRegistry
		wantErr bool
	}{
		{
			name:    "Positive createAnalysisTypesRegistry test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createAnalysisTypesRegistry()
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
