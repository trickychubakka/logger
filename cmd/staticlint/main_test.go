package main

import (
	"golang.org/x/tools/go/analysis"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// createConfigFile -- создание временного тестового конфиг файла.
func createConfigFile(filename string, badConfig bool, badFilePath bool) error {
	var json string
	json = `{"staticcheck": [
	"allSA"
	],
	"staticcheckexcl": [
	"SA1000",
	"SA1001"
	],
	"stylecheck": [
	"allST"
	],
	"stylecheckexcl": [
	"ST1022",
	"ST1023"
	],
	"analysis": [
	"appends",
	"asmdecl",
	"assign"
	],
	"analysisexcl": [
	"fieldalignment",
	"shadow"
	]
	}`
	if badConfig {
		json = "wrong config file text"
	}
	appfile, err := os.Executable()
	if err != nil {
		log.Println("createConfigFile: os.Executable() error", err)
		return err
	}

	path := filepath.Join(filepath.Dir(appfile), filename)
	log.Println("Write configfile with path", path)

	if !badFilePath {
		err = os.WriteFile(filepath.Join(filepath.Dir(appfile), filename), []byte(json), 0644)
		if err != nil {
			log.Println("createConfigFile: os.WriteFile error", err)
			return err
		}
	}
	return nil
}

// removeConfigFile -- удаление временного тестового конфиг файла.
func removeConfigFile(filename string) error {
	appfile, err := os.Executable()
	if err != nil {
		log.Println("removeConfigFile: os.Executable() error", err)
		return err
	}
	err = os.Remove(filepath.Join(filepath.Dir(appfile), filename))
	if err != nil {
		log.Println("createConfigFile: os.WriteFile error", err)
		return err
	}
	return nil
}

func TestChecksCreate(t *testing.T) {
	type args struct {
		cfg          ConfigData
		typeRegistry map[string]*analysis.Analyzer
	}
	typeRegistry, _ := createAnalysisTypesRegistry()
	tests := []struct {
		name    string
		args    args
		want    []*analysis.Analyzer
		wantErr bool
	}{
		{
			name: "Positive ChecksCreate test",
			args: args{
				cfg: ConfigData{
					Staticcheck:     []string{"allSA"},
					StylecheckExcl:  []string{"SA1000"},
					Stylecheck:      []string{"allST"},
					StaticcheckExcl: []string{"ST1022"},
					Analysis:        []string{"appends", "assign"},
					AnalysisExcl:    []string{"shadow"},
				},
				typeRegistry: typeRegistry,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ChecksCreate(tt.args.cfg, tt.args.typeRegistry)
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_readConfig(t *testing.T) {
	type args struct {
		configFile  string
		badConfig   bool
		badFilePath bool
	}
	tests := []struct {
		name    string
		args    args
		want    ConfigData
		wantErr bool
	}{
		{
			name: "Positive readConfig test",
			args: args{
				configFile:  "multichecker_test.json",
				badConfig:   false,
				badFilePath: false,
			},
			wantErr: false,
		},
		{
			name: "read config json.Unmarshall error test",
			args: args{
				configFile:  "multichecker_test.json",
				badConfig:   true,
				badFilePath: false,
			},
			wantErr: true,
		},
		{
			name: "wrong config file path error test",
			args: args{
				configFile:  "multichecker_test.json",
				badConfig:   false,
				badFilePath: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createConfigFile(tt.args.configFile, tt.args.badConfig, tt.args.badFilePath)
			defer removeConfigFile(tt.args.configFile)
			if err != nil {
				t.Errorf("readConfig() error = %v", err)
			}
			_, err = readConfig(tt.args.configFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
