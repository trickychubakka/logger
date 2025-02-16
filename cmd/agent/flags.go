package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"logger/config"
	"logger/internal/encryption"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// IsValidIP функция для проверки на то, что строка является валидным ip адресом.
func IsValidIP(ip string) bool {
	res := net.ParseIP(ip)
	return res != nil
}

// GetAgentOutboundIP get preferred outbound ip of this machine
// На хосте может быть несколько адресов. Ищем тот, который используется для исходящих в сторону сервера пакетов.
func GetAgentOutboundIP(extIP string) (string, error) {
	conn, err := net.Dial("udp", extIP)
	if err != nil {
		log.Println("GetAgentOutboundIP error :", err)
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

// initConfig функция инициализации конфигурации агента с использованием параметров командной строки.
func initConfig(conf *config.AgentConfig) error {
	var (
		ReportIntervalFlag string
		PollIntervalFlag   string
		AddressFlag        string
		LogFileFlag        string
		key                string
		RateLimitFlag      string
	)

	// Парсинг параметров командной строки.
	// Настройки переменных окружения имеют приоритет перед параметрами командной строки.
	if !FlagTest {
		flag.StringVar(&AddressFlag, "a", "localhost:8080", "address and port of logger server")
		//flag.StringVar(&AddressFlag, "a", "192.168.1.115:8080", "address and port of logger server")
		flag.StringVar(&ReportIntervalFlag, "r", "4", "agent report interval")
		flag.StringVar(&PollIntervalFlag, "p", "1", "agent poll interval")
		// Для логирования агента в лог файл необходимо определить флаг -l
		flag.StringVar(&LogFileFlag, "f", "", "agent log file")
		flag.StringVar(&key, "k", "", "key")
		//flag.StringVar(&key, "k", "superkey", "key")
		flag.StringVar(&RateLimitFlag, "l", "10", "Rate limit for agent connections to server.")
		flag.BoolVar(&conf.PProfHTTPEnabled, "t", false, "Flag for enabling pprof web server. Default false.")
		flag.StringVar(&conf.PathToPublicKey, "crypto-key", "./id_rsa.pub", "Path to public key. Default is ./id_rsa.pub")
		//flag.StringVar(&conf.PathToPublicKey, "crypto-key", "./id_rsa.pub", "Path to public key. Default is ./id_rsa.pub")
		flag.BoolVar(&conf.GRPCEnabled, "g", true, "Flag for enabling grpc. Default false.")
		flag.StringVar(&conf.GRPCRunAddr, "grpc-address", "localhost:3200", "address and port of gRPC server. Default localhost:3200.")

		flag.Parse()
	}
	// address processing.
	if envAddressFlag := os.Getenv("ADDRESS"); envAddressFlag != "" {
		log.Println("env var ADDRESS was specified, use ADDRESS =", envAddressFlag)
		AddressFlag = envAddressFlag
	}

	// Проверка на то, что заданный адрес является валидным IP или URI.
	if IsValidIP(strings.Split(AddressFlag, ":")[0]) {
		log.Println("AddressFlag is IP address, Using IP:", AddressFlag)
	} else if _, err := url.ParseRequestURI(AddressFlag); err != nil {
		log.Println("initConfig: Error parsing AddressFlag: ", AddressFlag)
		return err
	}
	conf.Address = AddressFlag

	// reportInterval processing.
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		log.Println("env var REPORT_INTERVAL was specified, use REPORT_INTERVAL =", envReportInterval)
		ReportIntervalFlag = envReportInterval
	}

	if c, err := strconv.Atoi(ReportIntervalFlag); err == nil {
		conf.ReportInterval = c
	} else {
		log.Println("initConfig: Error parsing ReportIntervalFlag: ", ReportIntervalFlag)
		return err
	}

	// PollInterval processing.
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		log.Println("env var POLL_INTERVAL was specified, use POLL_INTERVAL =", envPollInterval)
		PollIntervalFlag = envPollInterval
	}

	if c, err := strconv.Atoi(PollIntervalFlag); err == nil {
		conf.PollInterval = c
	} else {
		log.Println("initConfig: Error parsing PollIntervalFlag: ", PollIntervalFlag)
		return err
	}

	// pollInterval должен быть меньше, чем repInterval.
	if conf.PollInterval > conf.ReportInterval {
		return errors.New("poll interval must be less than report interval")
	}

	// LogFile processing.
	// Для логирования агента в лог файл необходимо определить переменную окружения AGENT_LOG.
	// Настройка переменных окружения имеют приоритет перед параметрами командной строки.
	if envLogFileFlag := os.Getenv("AGENT_LOG"); envLogFileFlag != "" {
		log.Println("env var AGENT_LOG was specified, use AGENT_LOG =", envLogFileFlag)
		LogFileFlag = envLogFileFlag
	}
	conf.Logfile = LogFileFlag

	if envKey := os.Getenv("KEY"); envKey != "" {
		log.Println("KEY env var specified")
		key = envKey
	}
	conf.Key = key

	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		log.Println("RATE_LIMIT env var specified, ", envRateLimit)
		RateLimitFlag = envRateLimit
	}
	if c, err := strconv.Atoi(RateLimitFlag); err == nil {
		conf.RateLimit = c
	} else {
		log.Println("initConfig: Error parsing RateLimitFlag: ", RateLimitFlag)
		return err
	}

	// Если CRYPTO_KEY определена -- переопределяем conf.PathToPrivateKey ее значением.
	if envPathToPublicKey := os.Getenv("CRYPTO_KEY"); envPathToPublicKey != "" {
		log.Println("env var CRYPTO_KEY defined, use CRYPTO_KEY value", envPathToPublicKey)
		conf.PathToPublicKey = envPathToPublicKey
	}

	if conf.PathToPublicKey != "" {
		publicKey, err := encryption.ReadPublicKeyFile(conf.PathToPublicKey)
		if err != nil {
			log.Println("InitConfig: Error reading public key file", err)
			return fmt.Errorf("%s %v", "Error reading public key file", err)
		}
		conf.PublicKey = publicKey
	}

	ip, err1 := GetAgentOutboundIP(conf.Address)
	if err1 != nil {
		log.Println("InitConfig: Error getting agent outbound IP address", err)
		return fmt.Errorf("%s %v", "Error getting agent outbound IP address", err)
	}
	conf.AgentIP = ip

	log.Printf("Agent starting with conf %s \n", fmt.Sprintf("%+v\n", conf))
	return nil
}
