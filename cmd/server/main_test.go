package main

import (
	"logger/cmd/server/initconf"
	"os"
	"testing"
)

// setEnv вспомогательная функция для установки переменных среды как параметров тестирования.
func setEnv(envAddr string) error {
	if err := os.Setenv("ADDRESS", envAddr); err != nil {
		return err
	}
	return nil
}

func Test_initConfig(t *testing.T) {

	type args struct {
		conf                initconf.Config
		envAddr             string
		StoreMetricInterval int
		FileStoragePath     string
		Restore             bool
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test initConfig",
			args: args{
				conf:    initconf.Config{RunAddr: "localhost:8080", FileStoragePath: "dump", Restore: true},
				envAddr: "localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "Negative Test initConfig, wrong URL",
			args: args{
				conf:    initconf.Config{RunAddr: "localhost:8080", FileStoragePath: "dump", Restore: true},
				envAddr: "d45656&&^%kjh",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		// Включение режима тестирования для отключения парсинга параметров командной строки
		initconf.FlagTest = true

		t.Run(tt.name, func(t *testing.T) {
			if err := setEnv(tt.args.envAddr); err != nil {
				panic(err)
			}
			//if err := initConfig(tt.args.h, tt.args.r, tt.args.p, &tt.args.conf); (err != nil) != tt.wantErr {
			if err := initconf.InitConfig(&tt.args.conf); (err != nil) != tt.wantErr {
				t.Errorf("initConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
