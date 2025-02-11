// Package conf -- пакет, содержащий объекты конфигураций logger.
package config

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
)

// DBConfig конфигурация базы данных logger сервера.
type DBConfig struct {
	Database struct {
		Host     string `mapstructure:"host"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Dbname   string `mapstructure:"dbname"`
		Sslmode  string `mapstructure:"sslmode"`
	}
}

// Config конфигурация logger сервера метрик.
type Config struct {
	RunAddr             string          `json:"address"`           // Address and port to run server.
	Logfile             string          `json:"logfile"`           // Server log file.
	StoreMetricInterval int             `json:"store_interval"`    // Store metrics dump to disk interval in sec.
	FileStoragePath     string          `json:"store_file"`        // File to save metrics to disk. For MemStorage type only.
	Restore             bool            `json:"restore"`           // Restore metrics dump with server start. For MemStorage type only.
	DatabaseDSN         string          `json:"database_dsn"`      // DatabaseDSN
	UseDBConfig         bool            `json:"use_db_config"`     // Use dbconfig/config yaml file (conf/dbconfig.yaml).
	Key                 string          `json:"key"`               // Key for HMAC.
	PProfHTTPEnabled    bool            `json:"pprof_http_enable"` // Start PProfHTTP server.
	TestDBMode          bool            `json:""`                  // Turn On test DB mode. For DB methods unit testing.
	TestMode            bool            `json:""`                  // Turn On test mode. For unit testing.
	PathToPrivateKey    string          `json:"crypto_key"`        // Path to private key.
	PrivateKey          *rsa.PrivateKey `json:""`                  // RSA private key.
	TrustedSubnet       string          `json:"trusted_subnet"`    // Trusted subnet in CIDR format.
}

// AgentConfig конфигурация logger агента.
type AgentConfig struct {
	PollInterval     int            `json:"poll_interval"`     // Agent metric polling interval.
	ReportInterval   int            `json:"report_interval"`   // Agent metric report interval.
	Address          string         `json:"address"`           // Logger server address and port.
	Logfile          string         `json:"agent_log"`         // Agent log file.
	Key              string         `json:""`                  // HMAC key.
	RateLimit        int            `json:"rate_limit"`        // Rate limit for agent connections to server.
	PProfHTTPEnabled bool           `json:"pprof_http_enable"` // Flag for enabling pprof web server.
	PathToPublicKey  string         `json:"crypto_key"`        // Path to RSA public key file.
	PublicKey        *rsa.PublicKey `json:""`                  // RSA public key.
	AgentIP          string         `json:"agent_ip"`          // Agent IP address (preferred outbound ip address).
}

// ToJSON конвертация JSON с go-style комментариями в "чистый" JSON для json.Unmarshal.
// Внимание! Комментарии в конце строки должны быть оформлены в виде " // ".
func ToJSON(b []byte) []byte {
	var res [][]byte
	for _, s := range bytes.Split(b, []byte("\n")) {
		// Комментарии с начала строки.
		if bytes.HasPrefix(bytes.TrimLeft(s, " "), []byte("//")) {
			continue
		}
		res = append(res, bytes.Split(s, []byte(" // "))[0])
	}
	return bytes.Join(res, []byte("\n"))
}

// ReadConfig чтение конфигурации из json файла.
// Внимание: чтение конфигурации создано в предположении, что значения интервалов в них задаются как int (в секундах),
// а не как string вида "1s" (см. задание по Инкремент 22) -- иначе будет необходима конвертация поля.
func ReadConfig(fileName string, conf any) error {
	var err error
	serverConf := Config{}
	agentConf := AgentConfig{}
	log.Println("ReadConfig: start to read config file", fileName)
	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Println("ReadConfig. Error in os.ReadFile() :", err)
		return err
	}
	jsn := ToJSON(data)
	log.Println("ReadConfig. jsn is: ", string(jsn))
	// Если на вход получена конфигурация сервера.
	if reflect.DeepEqual(reflect.TypeOf(conf), reflect.TypeOf(&serverConf)) {
		log.Println("reflect.TypeOf(conf) is *Config")
		err = json.Unmarshal(jsn, &conf)
		// Если на вход получена конфигурация агента.
	} else if reflect.TypeOf(conf) == reflect.TypeOf(&agentConf) {
		log.Println("reflect.TypeOf(conf) is *AgentConfig")
		err = json.Unmarshal(jsn, &conf)
		// Если полученное на вход непонятно.
	} else {
		return errors.New("wrong config")
	}
	if err != nil {
		log.Println("ReadConfig. Error in json.Unmarshal :", err)
		return err
	}
	log.Println("conf after Unmarshal is :", fmt.Sprintf("%+v\n", conf))
	return nil
}
