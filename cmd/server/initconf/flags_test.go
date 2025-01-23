package initconf

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"strings"
	"testing"
)

// setEnv вспомогательная функция для установки переменных среды как параметров тестирования.
func setEnv(envAddr string, envStoreInterval string, envRestore string, suffix string) error {
	if err := os.Setenv("ADDRESS"+suffix, envAddr); err != nil {
		return err
	}
	if err := os.Setenv("STORE_INTERVAL"+suffix, envStoreInterval); err != nil {
		return err
	}
	if err := os.Setenv("RESTORE"+suffix, envRestore); err != nil {
		return err
	}
	return nil
}

func createDBConfig(filename string, cfgContent string) error {
	testConfig := cfgContent

	// Создание файла с именем "greeting.txt" и запись в него данных
	err := os.WriteFile(filename, []byte(testConfig), 0644)
	if err != nil {
		log.Println("createDBConfig Error:", err)
		return err
	}
	return nil
}

// deleteDBConfig вспомогательная функция удаления временного файла конфигурации
func deleteDBConfig(filename string) error {
	if err := os.Remove(filename); err != nil {
		return err
	}
	return nil
}

// randomString генерация случайной строки заданной длины для префикса тестовых таблиц unit-тестов.
func randomString(length int) string {
	return uuid.NewString()[:length]
}

func Test_initConfig(t *testing.T) {

	type args struct {
		conf                Config
		envAddr             string
		envStoreInterval    string
		envRestore          string
		StoreMetricInterval int
		FileStoragePath     string
		Restore             bool
	}
	testSuffix := strings.ToUpper(fmt.Sprintf("_%s", randomString(4)))
	log.Println("testSuffix is :", testSuffix)
	suffix = testSuffix

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test initConfig",
			args: args{
				conf:    Config{RunAddr: "localhost:8080", FileStoragePath: "dump", Restore: true, TestMode: true},
				envAddr: "localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "Negative Test initConfig, does not match the pattern address:port",
			args: args{
				conf:    Config{RunAddr: "localhost:8080", FileStoragePath: "dump", Restore: true, TestMode: true},
				envAddr: "d45656&&^%kjh",
			},
			wantErr: true,
		},
		{
			name: "Negative Test initConfig, port is not number",
			args: args{
				conf:    Config{RunAddr: "localhost:8080", FileStoragePath: "dump", Restore: true, TestMode: true},
				envAddr: "localhost:wrongPort",
			},
			wantErr: true,
		},
		{
			name: "Negative Test initConfig, wrong RequestURI",
			args: args{
				conf:    Config{RunAddr: "localhost:8080", FileStoragePath: "dump", Restore: true, TestMode: true},
				envAddr: "localhost;:9090",
			},
			wantErr: true,
		},
		{
			name: "Negative Test initConfig, invalid STORE_INTERVAL variable",
			args: args{
				conf:             Config{RunAddr: "localhost:8080", FileStoragePath: "dump", Restore: true, TestMode: true},
				envAddr:          "localhost:8080",
				envStoreInterval: "WrongStoreInterval",
			},
			wantErr: true,
		},
		{
			name: "Negative Test initConfig, invalid RESTORE variable",
			args: args{
				conf:             Config{RunAddr: "localhost:8080", FileStoragePath: "dump", TestMode: true},
				envAddr:          "localhost:8080",
				envRestore:       "TRUGH",
				envStoreInterval: "WrongStoreInterval",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		// Включение режима тестирования для отключения парсинга параметров командной строки
		FlagTest = true

		t.Run(tt.name, func(t *testing.T) {
			if err := setEnv(tt.args.envAddr, tt.args.envStoreInterval, tt.args.envRestore, suffix); err != nil {
				panic(err)
			}
			if err := InitConfig(&tt.args.conf); err != nil {
				if !tt.wantErr {
					t.Errorf("initConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
				log.Println("initConfig() error:", err)
			}
		})
	}
	//for _, e := range os.Environ() {
	//	fmt.Println(e)
	//}
}

func TestIsValidIP(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Positive Test TestIsValidIP",
			args: args{
				ip: "192.168.1.1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidIP(tt.args.ip); got != tt.want {
				t.Errorf("IsValidIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readDBConfig(t *testing.T) {
	type args struct {
		path       string
		filename   string
		cfgContent string
	}
	tests := []struct {
		name string
		args args
		//want    string
		wantErr bool
	}{
		{
			name: "Positive Test Test_readDBConfig",
			args: args{
				path:     `.`,
				filename: "TestDBConfig.yaml",
				cfgContent: `database:
  host: "192.168.1.100"
  user: "testuser"
  password: "123456"
  dbname: "testdb"
  sslmode: "disable"`,
			},
			wantErr: false,
		},
		{
			name: "Error Read config file Test_readDBConfig",
			args: args{
				path:     `.`,
				filename: "",
				cfgContent: `database:
  host: "192.168.1.100"
  user: "testuser"
  password: "123456"
  dbname: "testdb"
  sslmode: "disable"`,
			},
			wantErr: true,
		},
		{
			name: "Error unmarshalling conf Test_readDBConfig",
			args: args{
				path:       `.`,
				filename:   "TestDBConfig.yaml",
				cfgContent: `Incorrect file`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаем временный файл конфигурации
			err := createDBConfig(tt.args.filename, tt.args.cfgContent)
			if err != nil {
				log.Println("createDBConfig Error:", err)
			}
			_, err = readDBConfig(tt.args.filename, tt.args.path)
			pathToDelete := fmt.Sprintf("%s\\%s", tt.args.path, tt.args.filename)
			log.Println("pathToDelete:", pathToDelete)
			// удаляем временный файл конфигурации
			defer deleteDBConfig("TestDBConfig.yaml")
			if (err != nil) != tt.wantErr {
				t.Errorf("readDBConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
