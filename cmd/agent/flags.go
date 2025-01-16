package main

import (
	"errors"
	"flag"
	"log"
	"logger/conf"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var config conf.AgentConfig

// IsValidIP функция для проверки на то, что строка является валидным ip адресом
func IsValidIP(ip string) bool {
	res := net.ParseIP(ip)
	return res != nil
}

// initConfig функция инициализации конфигурации агента с использованием параметров командной строки
func initConfig(conf *conf.AgentConfig) error {

	var (
		ReportIntervalFlag string
		PollIntervalFlag   string
		AddressFlag        string
		LogFileFlag        string
		key                string
		RateLimitFlag      string
	)

	// Парсинг параметров командной строки
	// Настройка переменных окружения имеют приоритет перед параметрами командной строки
	if !FlagTest {
		flag.StringVar(&AddressFlag, "a", "localhost:8080", "address and port to run server")
		flag.StringVar(&ReportIntervalFlag, "r", "4", "agent report interval")
		flag.StringVar(&PollIntervalFlag, "p", "1", "agent poll interval")
		// Для логирования агента в лог файл необходимо определить флаг -l
		flag.StringVar(&LogFileFlag, "f", "", "agent log file")
		flag.StringVar(&key, "k", "", "key")
		flag.StringVar(&RateLimitFlag, "l", "10", "Rate limit for agent connections to server.")
		flag.BoolVar(&conf.PProfHTTPEnabled, "t", true, "Flag for enabling pprof web server. Default false.")

		flag.Parse()
	}
	// address processing
	if envAddressFlag := os.Getenv("ADDRESS"); envAddressFlag != "" {
		log.Println("env var ADDRESS was specified, use ADDRESS =", envAddressFlag)
		AddressFlag = envAddressFlag
	}

	// Проверка на то, что заданный адрес является валидным IP или URI
	if IsValidIP(strings.Split(AddressFlag, ":")[0]) {
		log.Println("AddressFlag is IP address, Using IP:", AddressFlag)
	} else if _, err := url.ParseRequestURI(AddressFlag); err != nil {
		return err
	}
	conf.Address = AddressFlag

	// reportInterval processing
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		log.Println("env var REPORT_INTERVAL was specified, use REPORT_INTERVAL =", envReportInterval)
		ReportIntervalFlag = envReportInterval
	}

	if c, err := strconv.Atoi(ReportIntervalFlag); err == nil {
		conf.ReportInterval = c
	} else {
		return err
	}

	// PollInterval processing
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		log.Println("env var POLL_INTERVAL was specified, use POLL_INTERVAL =", envPollInterval)
		PollIntervalFlag = envPollInterval
	}

	if c, err := strconv.Atoi(PollIntervalFlag); err == nil {
		conf.PollInterval = c
	} else {
		return err
	}

	// pollInterval должен быть меньше, чем repInterval
	if conf.PollInterval > conf.ReportInterval {
		return errors.New("poll interval must be less than report interval")
	}

	// LogFile processing
	// Для логирования агента в лог файл необходимо определить переменную окружения AGENT_LOG
	// Настройка переменных окружения имеют приоритет перед параметрами командной строки
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
		log.Println("KEY env var specified, ", envRateLimit)
		RateLimitFlag = envRateLimit
	}
	if c, err := strconv.Atoi(RateLimitFlag); err == nil {
		conf.RateLimit = c
	} else {
		return err
	}

	log.Printf("Address is %s, PollInterval is %d, ReportInterval is %d, LogFile is %s, RateLimit id %d \n", conf.Address, conf.PollInterval, conf.ReportInterval, conf.Logfile, conf.RateLimit)
	return nil
}
