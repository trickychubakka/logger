// Package initconf -- пакет инициализации конфигурации logger сервера.
package initconf

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"logger/config"
	"logger/internal/encryption"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config объект конфигурации logger сервера метрик.
//type Config struct {
//	RunAddr             string          // Address and port to run server.
//	Logfile             string          // Server log file.
//	StoreMetricInterval int             // Store metrics dump to disk interval in sec.
//	FileStoragePath     string          // File to save metrics to disk. For MemStorage type only.
//	Restore             bool            // Restore metrics dump with server start. For MemStorage type only.
//	DatabaseDSN         string          // DatabaseDSN
//	UseDBConfig         bool            // Use dbconfig/config yaml file (conf/dbconfig.yaml).
//	Key                 string          // Key for HMAC.
//	PProfHTTPEnabled    bool            // Start PProfHTTP server.
//	TestDBMode          bool            // Turn On test DB mode. For DB methods unit testing.
//	TestMode            bool            // Turn On test mode. For unit testing.
//	PathToPrivateKey    string          // Path to private key.
//	PrivateKey          *rsa.PrivateKey // RSA private key.
//}

// IsValidIP функция для проверки на то, что строка является валидным ip адресом.
func IsValidIP(ip string) bool {
	res := net.ParseIP(ip)
	return res != nil
}

// FlagTest флаг режима тестирования для отключения парсинга командной строки при тестировании.
var FlagTest = false

// suffix суффикс для переменных окружения при тестировании. Для предотвращения затирания переменных окружения.
var suffix = ""

// readDBConfig -- функция чтения конфигурации dbconfig.yaml по указанному пути.
func readDBConfig(configName string, configPath string) (string, error) {
	dbCfg := &config.DBConfig{}
	var connStr string
	log.Println("flags and DATABASE_DSN env are not defined, trying to find and read dbconfig.yaml")
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Error reading conf file :", err)
		return "", err
	}
	err2 := viper.Unmarshal(&dbCfg)
	if err2 != nil {
		log.Println("Error unmarshalling conf :", err2)
		return "", err2
	}
	connStr = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=%s", dbCfg.Database.User, dbCfg.Database.Password, dbCfg.Database.Host, dbCfg.Database.Dbname, dbCfg.Database.Sslmode)
	return connStr, nil
}

// InitConfig -- функция инициализации конфигурации logger сервера.
// Конфигурируемые параметры определяется через параметры запуска командной строки либо через переменные окружения.
// Переменные окружения имеют приоритет перед параметрами командной строки.
func InitConfig(conf *config.Config) error {
	var envGenerateRSAKeys bool
	var err error
	log.Println("InitConfig, SUFFIX =", suffix)
	if !FlagTest {
		log.Println("start parsing flags")
		flag.StringVar(&conf.RunAddr, "a", "localhost:8080", "address and port to run server. Default localhost:8080.")
		flag.StringVar(&conf.Logfile, "l", "", "server log file. Default empty.")
		flag.IntVar(&conf.StoreMetricInterval, "i", 10, "store metrics to disk interval in sec. 0 -- sync saving. Default 10 sec.")
		flag.StringVar(&conf.FileStoragePath, "f", "metrics.dump", "file to save metrics to disk. Default metric_dump.json.")
		flag.BoolVar(&conf.Restore, "r", true, "true/false flag -- restore metrics dump with server start. Default true.")
		flag.StringVar(&conf.DatabaseDSN, "d", "", "database DSN in format postgres://user:password@host:port/dbname?sslmode=disable. Default is empty.")
		//flag.StringVar(&conf.DatabaseDSN, "d", "postgres://testuser:123456@192.168.1.100:5432/testdb?sslmode=disable", "database DSN in format postgres://user:password@host:port/dbname?sslmode=disable. Default is empty.")
		//flag.StringVar(&conf.Key, "k", "", "Key. Default empty.")
		flag.StringVar(&conf.Key, "k", "superkey", "Key. Default empty.")
		flag.StringVar(&conf.PathToPrivateKey, "crypto-key", "./id_rsa", "Path to private key. Default is ./id_rsa")
		//flag.StringVar(&conf.PathToPrivateKey, "crypto-key", "", "Path to private key. Default is ./id_rsa")
		flag.BoolVar(&envGenerateRSAKeys, "generate-keys", true, "To generation new RSA public and private "+
			"keys and save them to the same with conf.PathToPrivateKey directory. Default false. "+
			"Naming: private key filename defined with -generate-keys option, public filename will be `privateKey filename + .pub`")
		flag.BoolVar(&conf.UseDBConfig, "c", false, "true/false flag -- use dbconfig/config yaml file +(conf/dbconfig.yaml). Default false.")
		flag.BoolVar(&conf.PProfHTTPEnabled, "t", true, "Flag for enabling pprof web server. Default false.")
		flag.Parse()
	}

	log.Println("Config after flag but before env var processing:", fmt.Sprintf("%+v\n", conf))

	// Пытаемся прочитать переменную окружения ADDRESS. Переменные окружения имеют приоритет перед флагами,
	// поэтому переопределяют опции командной строки в случае, если соответствующая переменная определена в env.
	log.Println("Trying to read ADDRESS environment variable (env has priority over flags): ", os.Getenv("ADDRESS"))
	if envRunAddr := os.Getenv("ADDRESS" + suffix); envRunAddr != "" {
		fmt.Println("Using env var ADDRESS:", envRunAddr)
		conf.RunAddr = envRunAddr
	}

	// Проверка на то, что заданный адрес является валидным сочетанием IP:порт.
	ipPort := strings.Split(conf.RunAddr, ":")
	// адрес состоит из сочетания хост:порт
	if len(ipPort) != 2 || ipPort[1] == "" {
		return fmt.Errorf("invalid ADDRESS variable `%s`, does not match the pattern address:port ", conf.RunAddr)
	}
	// Порт содержит только цифры.
	if _, err := strconv.Atoi(ipPort[1]); err != nil {
		return fmt.Errorf("invalid ADDRESS variable `%s`, port is not number", conf.RunAddr)
	}
	// Если часть URI является валидным IP.
	if IsValidIP(ipPort[0]) {
		log.Println("conf.runAddr is IP address, Using IP:", conf.RunAddr)
		//return nil
	}
	// Если адрес не является валидным URI -- возвращаем ошибку.
	if _, err := url.ParseRequestURI(conf.RunAddr); err != nil {
		log.Println("Error parsing RequestURI", err)
		return fmt.Errorf("invalid ADDRESS variable `%s`, wrong RequestURI", conf.RunAddr)
	}

	if envLogFileFlag := os.Getenv("SERVER_LOG" + suffix); envLogFileFlag != "" {
		log.Println("env var SERVER_LOG was specified, use SERVER_LOG =", envLogFileFlag)
		conf.Logfile = envLogFileFlag
		log.Println("Using env var SERVER_LOG=", envLogFileFlag)
	}

	if envStoreMetricInterval := os.Getenv("STORE_INTERVAL" + suffix); envStoreMetricInterval != "" {
		log.Println("env var STORE_INTERVAL was specified, use STORE_INTERVAL =", envStoreMetricInterval)
		tmp, err := strconv.Atoi(envStoreMetricInterval)
		if err != nil {
			return fmt.Errorf("invalid STORE_INTERVAL variable `%d`", tmp)
		}
		conf.StoreMetricInterval = tmp
		log.Println("Using env var STORE_INTERVAL=", conf.StoreMetricInterval)
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH" + suffix); envFileStoragePath != "" {
		log.Println("env var FILE_STORAGE_PATH was specified, use FILE_STORAGE_PATH =", envFileStoragePath)
		conf.FileStoragePath = envFileStoragePath
		log.Println("Using env var FILE_STORAGE_PATH=", conf.FileStoragePath)
	}

	if envRestore := os.Getenv("RESTORE" + suffix); envRestore != "" {
		log.Println("env var RESTORE was specified, use RESTORE =", envRestore)
		tmp, err := strconv.ParseBool(envRestore)
		if err != nil {
			return fmt.Errorf("invalid RESTORE variable `%t`", tmp)
		}
		conf.Restore = tmp
		log.Println("Using env var RESTORE=", conf.Restore)
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN" + suffix); envDatabaseDSN != "" {
		log.Println("env var DATABASE_DSN was specified, use DATABASE_DSN =", envDatabaseDSN)
		conf.DatabaseDSN = envDatabaseDSN
		log.Println("Using env var DATABASE_DSN=", conf.DatabaseDSN)
	}

	// Если DatabaseDSN нет в переменных окружения и в параметрах запуска -- пытаемся прочитать из dbconfig.yaml.
	if conf.DatabaseDSN == "" && conf.UseDBConfig {
		log.Println("flags and DATABASE_DSN env are not defined, trying to find and read dbconfig.yaml")
		if connStr, err := readDBConfig("dbconfig", "./conf"); err != nil {
			log.Println("Error reading dbconfig.yaml:", err)
		} else {
			conf.DatabaseDSN = connStr
		}
	}

	if envKey := os.Getenv("KEY" + suffix); envKey != "" {
		log.Println("env var KEY was specified, use KEY")
		conf.Key = envKey
		log.Println("Using key")
	}

	if envPathToPrivateKey := os.Getenv("CRYPTO_KEY" + suffix); envPathToPrivateKey != "" {
		log.Println("env var CRYPTO_KEY was specified, use CRYPTO_KEY in", envPathToPrivateKey)
		conf.PathToPrivateKey = envPathToPrivateKey
		log.Println("Using key", conf.PathToPrivateKey)
	} else {
		log.Println("env var CRYPTO_KEY not specified, use command line value from flag.StringVar(&conf.PathToPrivateKey...)", conf.PathToPrivateKey)
		log.Println("Using key", conf.PathToPrivateKey)
	}

	// Если CRYPTO_KEY определена -- переопределяем conf.PathToPrivateKey ее значением.
	if envPathToPrivateKey := os.Getenv("CRYPTO_KEY" + suffix); envPathToPrivateKey != "" {
		log.Println("env var CRYPTO_KEY defined, use CRYPTO_KEY value", envPathToPrivateKey)
		conf.PathToPrivateKey = envPathToPrivateKey
	}

	// Если определена опция generate-keys -- генерируются новые приватный и публичный ключи,
	// приватный сохраняется в conf.PrivateKey, ключи сохраняются в соответствующие файлы.
	// Filepath этих файлов определяется через conf.PathToPrivateKey
	if envGenerateRSAKeys {
		log.Println("Start to generate new RSA private and public keys")
		if conf.PathToPrivateKey == "" {
			log.Println("Error. conf.PathToPrivateKey (-crypto-key option or CRYPTO_KEY env var) must not be empty")
			return fmt.Errorf("%s", "Error, conf.PathToPrivateKey must not be empty")
		}
		privKeyFile := conf.PathToPrivateKey
		pubKeyFile := privKeyFile + ".pub"
		conf.PrivateKey, _, err = encryption.GenerateRSAKeyPair(privKeyFile, pubKeyFile)
		if err != nil {
			log.Println("Error generating RSA private key:", err)
			return err
		}
	}
	// Если определена опция или переменная окружения PathToPrivateKey и не включена опция generate-keys,
	// читаем приватный ключ из этого файла.
	if conf.PathToPrivateKey != "" && !envGenerateRSAKeys {
		privateKey, err := encryption.ReadPrivateKeyFile(conf.PathToPrivateKey)
		if err != nil {
			log.Println("InitConfig: Error reading private key file", err)
			return fmt.Errorf("%s %v", "Error reading private key file", err)
		}
		conf.PrivateKey = privateKey
	}

	log.Println("conf.runAddr is URI address, Using URI:", conf.RunAddr)
	return nil
}
