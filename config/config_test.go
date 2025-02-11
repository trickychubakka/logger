package config

import (
	"log"
	"os"
	"reflect"
	"testing"
)

func createTMPConfig(filename string, badJSON bool) error {
	var confString string
	confString = `{
  "address": "\"localhost:8080\"", // address and port of logger server. Аналог переменной окружения ADDRESS или флага -a
  "report_interval": 4, // agent report interval. Аналог переменной окружения REPORT_INTERVAL или флага -r
  "poll_interval": 1, // agent poll interval. Аналог переменной окружения POLL_INTERVAL или флага -p
  "crypto_key": "", // Path to public key. Аналог переменной окружения CRYPTO_KEY или флага -crypto-key
  "agent_log": "./id_rsa", // agent log file. Аналог переменной окружения AGENT_LOG или флага -f
  "rate_limit": 10, // Rate limit for agent connections to server. Аналог переменной окружения RATE_LIMIT или флага -l
  "pprof_http_enable": false // Flag for enabling pprof web server.
}`
	if badJSON {
		log.Println("BAD config")
		confString = `{
  "address": "localhost:8080", // address and port of logger server. Аналог переменной окружения ADDRESS или флага -a
  "report_interval": "4", // agent report interval. Аналог переменной окружения REPORT_INTERVAL или флага -r
  "poll_interval": "1", // agent poll interval. Аналог переменной окружения POLL_INTERVAL или флага -p
  "crypto_key": "", // Path to public key. Аналог переменной окружения CRYPTO_KEY или флага -crypto-key
  "agent_log": false, // agent log file. Аналог переменной окружения AGENT_LOG или флага -f
  "rate_limit": "test", // Rate limit for agent connections to server. Аналог переменной окружения RATE_LIMIT или флага -l
  "pprof_http_enable": false // Flag for enabling pprof web server.
}`
	}
	err := os.WriteFile(filename, []byte(confString), 0644)
	if err != nil {
		return err
	}
	return nil
}

func TestReadConfig(t *testing.T) {
	type args struct {
		fileName  string
		conf      any
		ErrorType string
		badJSON   bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive test ReadConfig",
			args: args{
				fileName: "./tempconfig11111.json",
				conf:     &AgentConfig{},
			},
			wantErr: false,
		},
		{
			name: "os.ReadFile error test ReadConfig",
			args: args{
				fileName:  "./tempconfig11111.json",
				conf:      &AgentConfig{},
				ErrorType: "os.ReadFile error",
			},
			wantErr: true,
		},
		{
			name: "os.ReadFile error test ReadConfig",
			args: args{
				fileName:  "./tempconfig11111.json",
				conf:      &AgentConfig{},
				ErrorType: "Error in json.Unmarshal",
				badJSON:   true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createTMPConfig(tt.args.fileName, tt.args.badJSON); err != nil {
				t.Errorf("createTMPConfig() error = %v", err)
			}
			if tt.args.ErrorType == "os.ReadFile error" {
				_ = os.Remove(tt.args.fileName)
			}
			if err := ReadConfig(tt.args.fileName, tt.args.conf); (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			_ = os.Remove(tt.args.fileName)
		})
	}
}

func TestToJSON(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Positive test ToJSON",
			args: args{
				s: `{
  "address": "localhost:8080", // address and port of logger server. Аналог переменной окружения ADDRESS или флага -a
  // COMMENT
  "report_interval": 4, // agent report interval. Аналог переменной окружения REPORT_INTERVAL или флага -r
  "poll_interval": 1, // agent poll interval. Аналог переменной окружения POLL_INTERVAL или флага -p
  "crypto_key": "", // Path to public key. Аналог переменной окружения CRYPTO_KEY или флага -crypto-key
  "agent_log": "./id_rsa", // agent log file. Аналог переменной окружения AGENT_LOG или флага -f
  "rate_limit": 10, // Rate limit for agent connections to server. Аналог переменной окружения RATE_LIMIT или флага -l
  "pprof_http_enable": false // Flag for enabling pprof web server.
}`,
			},
			want: `{
  "address": "localhost:8080",
  "report_interval": 4,
  "poll_interval": 1,
  "crypto_key": "",
  "agent_log": "./id_rsa",
  "rate_limit": 10,
  "pprof_http_enable": false
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToJSON([]byte(tt.args.s)); !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("ToJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
