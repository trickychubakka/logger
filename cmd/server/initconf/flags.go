package initconf

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"logger/conf"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	RunAddr             string
	Logfile             string
	StoreMetricInterval int
	FileStoragePath     string
	Restore             bool
	DatabaseDSN         string
	UseDBConfig         bool
	Key                 string
	PProfHTTPEnabled    bool
}

// IsValidIP функция для проверки на то, что строка является валидным ip адресом
func IsValidIP(ip string) bool {
	res := net.ParseIP(ip)
	return res != nil
}

// FlagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
var FlagTest = false

func readDBConfig() (string, error) {
	dbCfg := &conf.Config{}
	var connStr string
	log.Println("flags and DATABASE_DSN env are not defined, trying to find and read dbconfig.yaml")
	viper.SetConfigName("dbconfig")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./conf")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Error reading conf file :", err)
		return "", err
	} else {
		err = viper.Unmarshal(&dbCfg)
		if err != nil {
			log.Println("Error unmarshalling conf :", err)
			return "", err
		}
	}

	connStr = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=%s", dbCfg.Database.User, dbCfg.Database.Password, dbCfg.Database.Host, dbCfg.Database.Dbname, dbCfg.Database.Sslmode)
	return connStr, nil
}

func InitConfig(conf *Config) error {

	if !FlagTest {
		log.Println("start parsing flags")
		flag.StringVar(&conf.RunAddr, "a", "localhost:8080", "address and port to run server. Default localhost:8080.")
		flag.StringVar(&conf.Logfile, "l", "", "server log file. Default empty.")
		flag.IntVar(&conf.StoreMetricInterval, "i", 10, "store metrics to disk interval in sec. 0 -- sync saving. Default 10 sec.")
		flag.StringVar(&conf.FileStoragePath, "f", "metrics.dump", "file to save metrics to disk. Default metric_dump.json.")
		flag.BoolVar(&conf.Restore, "r", true, "true/false flag -- restore metrics dump with server start. Default true.")
		//flag.StringVar(&conf.DatabaseDSN, "d", "", "database DSN in format postgres://user:password@host:port/dbname?sslmode=disable. Default is empty.")
		flag.StringVar(&conf.DatabaseDSN, "d", "postgres://testuser:123456@192.168.1.100:5432/testdb?sslmode=disable", "database DSN in format postgres://user:password@host:port/dbname?sslmode=disable. Default is empty.")
		flag.StringVar(&conf.Key, "k", "", "Key. Default empty.")
		flag.BoolVar(&conf.UseDBConfig, "c", false, "true/false flag -- use dbconfig/config yaml file (conf/dbconfig.yaml). Default false.")
		flag.BoolVar(&conf.PProfHTTPEnabled, "t", true, "Flag for enabling pprof web server. Default false.")
		flag.Parse()
	}

	log.Println("Config before env var processing:", conf)

	// Пытаемся прочитать переменную окружения ADDRESS. Переменные окружения имеют приоритет перед флагами,
	// поэтому переопределяют опции командной строки в случае, если соответствующая переменная определена в env
	log.Println("Trying to read ADDRESS environment variable (env has priority over flags): ", os.Getenv("ADDRESS"))
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		fmt.Println("Using env var ADDRESS:", envRunAddr)
		conf.RunAddr = envRunAddr
	}

	// Проверка на то, что заданный адрес является валидным сочетанием IP:порт
	ipPort := strings.Split(conf.RunAddr, ":")
	// адрес состоит из сочетания хост:порт
	if len(ipPort) != 2 || ipPort[1] == "" {
		return fmt.Errorf("invalid ADDRESS variable `%s`", conf.RunAddr)
	}
	// Порт содержит только цифры
	if _, err := strconv.Atoi(ipPort[1]); err != nil {
		return fmt.Errorf("invalid ADDRESS variable `%s`", conf.RunAddr)
	}
	// Если часть URI является валидным IP
	if IsValidIP(ipPort[0]) {
		log.Println("conf.runAddr is IP address, Using IP:", conf.RunAddr)
		return nil
	}
	// Если адрес не является валидным URI -- возвращаем ошибку
	if _, err := url.ParseRequestURI(conf.RunAddr); err != nil {
		log.Println("Error parsing RequestURI", err)
		return fmt.Errorf("invalid ADDRESS variable `%s`", conf.RunAddr)
	}

	if envLogFileFlag := os.Getenv("SERVER_LOG"); envLogFileFlag != "" {
		log.Println("env var SERVER_LOG was specified, use SERVER_LOG =", envLogFileFlag)
		conf.Logfile = envLogFileFlag
		log.Println("Using env var SERVER_LOG=", envLogFileFlag)
	}

	if envStoreMetricInterval := os.Getenv("STORE_INTERVAL"); envStoreMetricInterval != "" {
		log.Println("env var STORE_INTERVAL was specified, use STORE_INTERVAL =", envStoreMetricInterval)
		tmp, err := strconv.Atoi(envStoreMetricInterval)
		if err != nil {
			return fmt.Errorf("invalid STORE_INTERVAL variable `%d`", tmp)
		}
		conf.StoreMetricInterval = tmp
		log.Println("Using env var STORE_INTERVAL=", conf.StoreMetricInterval)
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		log.Println("env var FILE_STORAGE_PATH was specified, use FILE_STORAGE_PATH =", envFileStoragePath)
		conf.FileStoragePath = envFileStoragePath
		log.Println("Using env var FILE_STORAGE_PATH=", conf.FileStoragePath)
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		log.Println("env var RESTORE was specified, use RESTORE =", envRestore)
		tmp, err := strconv.ParseBool(envRestore)
		if err != nil {
			return fmt.Errorf("invalid RESTORE variable `%t`", tmp)
		}
		conf.Restore = tmp
		log.Println("Using env var RESTORE=", conf.Restore)
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		log.Println("env var DATABASE_DSN was specified, use DATABASE_DSN =", envDatabaseDSN)
		conf.DatabaseDSN = envDatabaseDSN
		log.Println("Using env var DATABASE_DSN=", conf.DatabaseDSN)
	}

	// Если DatabaseDSN нет в переменных окружения и в параметрах запуска -- пытаемся прочитать из dbconfig.yaml
	if conf.DatabaseDSN == "" && conf.UseDBConfig {
		log.Println("flags and DATABASE_DSN env are not defined, trying to find and read dbconfig.yaml")
		if connStr, err := readDBConfig(); err != nil {
			log.Println("Error reading dbconfig.yaml:", err)
		} else {
			conf.DatabaseDSN = connStr
		}
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		log.Println("env var DATABASE_DSN was specified, use DATABASE_DSN =", envKey)
		conf.Key = envKey
		log.Println("Using key")
	}

	log.Println("conf.runAddr is URI address, Using URI:", conf.RunAddr)
	return nil
}
