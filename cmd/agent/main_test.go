package main

import (
	"logger/conf"
	"os"
	"testing"
)

// setEnv вспомогательная функция для установки переменных среды как параметров тестирования
func setEnv(envAddr, envPollInterval, envReportInterval string) error {
	if err := os.Setenv("ADDRESS", envAddr); err != nil {
		return err
	}
	if err := os.Setenv("POLL_INTERVAL", envPollInterval); err != nil {
		return err
	}
	if err := os.Setenv("REPORT_INTERVAL", envReportInterval); err != nil {
		return err
	}
	if err := os.Setenv("RATE_LIMIT", envReportInterval); err != nil {
		return err
	}
	return nil
}

func Test_initConfig(t *testing.T) {

	type args struct {
		conf              conf.AgentConfig
		envAddr           string
		envPollInterval   string
		envReportInterval string
		envRateLimit      string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test initConfig",
			args: args{
				conf:              conf.AgentConfig{PollInterval: 10, ReportInterval: 2, Address: "localhost:8080"},
				envAddr:           "localhost:8080",
				envPollInterval:   "2",
				envReportInterval: "10",
				envRateLimit:      "5",
			},
			wantErr: false,
		},
		{
			name: "Negative Test initConfig, wrong URL",
			args: args{
				conf:              conf.AgentConfig{PollInterval: 10, ReportInterval: 2, Address: "localhost:8080"},
				envAddr:           "d45656&&^%kjh",
				envPollInterval:   "2",
				envReportInterval: "10",
			},
			wantErr: true,
		},
		{
			name: "Negative Test initConfig, wrong reportInterval",
			args: args{
				conf:              conf.AgentConfig{},
				envAddr:           "localhost:8080",
				envPollInterval:   "2",
				envReportInterval: "ere",
			},
			wantErr: true,
		},
		{
			name: "Negative Test initConfig, wrong pollingInterval",
			args: args{
				conf:              conf.AgentConfig{},
				envAddr:           "localhost:8080",
				envPollInterval:   "ere",
				envReportInterval: "10",
			},
			wantErr: true,
		},
		{
			name: "Negative Test initConfig, poll interval must be less than report interval",
			args: args{
				conf:              conf.AgentConfig{},
				envAddr:           "localhost:7777",
				envPollInterval:   "20",
				envReportInterval: "10",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		// Включение режима тестирования для отключения парсинга параметров командной строки
		FlagTest = true

		t.Run(tt.name, func(t *testing.T) {
			if err := setEnv(tt.args.envAddr, tt.args.envPollInterval, tt.args.envReportInterval); err != nil {
				panic(err)
			}
			//if err := initConfig(tt.args.h, tt.args.r, tt.args.p, &tt.args.conf); (err != nil) != tt.wantErr {
			if err := initConfig(&tt.args.conf); (err != nil) != tt.wantErr {
				t.Errorf("initConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
