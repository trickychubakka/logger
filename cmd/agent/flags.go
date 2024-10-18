package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	pollInterval   int
	reportInterval int
	address        string
	logfile        string
}

// IsValidIP функция для проверки на то, что строка является валидным ip адресом
func IsValidIP(ip string) bool {
	res := net.ParseIP(ip)
	return res != nil
	//if res == nil {
	//	return false
	//}
	//return true
}

// initConfig функция инициализации конфигурации агента с использованием параметров командной строки
func initConfig(conf *Config) error {

	var (
		ReportIntervalFlag string
		PollIntervalFlag   string
		AddressFlag        string
		LogFileFlag        string
	)

	// Парсинг параметров командной строки
	// Настройка переменных окружения имеют приоритет перед параметрами командной строки
	if !FlagTest {
		flag.StringVar(&AddressFlag, "a", "localhost:8080", "address and port to run server")
		flag.StringVar(&ReportIntervalFlag, "r", "10", "agent report interval")
		flag.StringVar(&PollIntervalFlag, "p", "2", "agent poll interval")
		// Для логирования агента в лог файл необходимо определеить флаг -l
		flag.StringVar(&LogFileFlag, "l", "", "agent log file")

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
	conf.address = AddressFlag

	// reportInterval processing
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		log.Println("env var REPORT_INTERVAL was specified, use REPORT_INTERVAL =", envReportInterval)
		ReportIntervalFlag = envReportInterval
	}

	if c, err := strconv.Atoi(ReportIntervalFlag); err == nil {
		conf.reportInterval = c
	} else {
		return err
	}

	// PollInterval processing
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		log.Println("env var POLL_INTERVAL was specified, use POLL_INTERVAL =", envPollInterval)
		PollIntervalFlag = envPollInterval
	}

	if c, err := strconv.Atoi(PollIntervalFlag); err == nil {
		conf.pollInterval = c
	} else {
		return err
	}

	// pollInterval должен быть меньше, чем repInterval
	if conf.pollInterval > conf.reportInterval {
		return errors.New("poll interval must be less than report interval")
	}

	// LogFile processing
	// Для логирования агента в лог файл необходимо определеить переменную окружения AGENT_LOG
	// Настройка переменных окружения имеют приоритет перед параметрами командной строки
	if envLogFileFlag := os.Getenv("AGENT_LOG"); envLogFileFlag != "" {
		log.Println("env var AGENT_LOG was specified, use AGENT_LOG =", envLogFileFlag)
		LogFileFlag = envLogFileFlag
	}
	conf.logfile = LogFileFlag

	log.Printf("Address is %s, PollInterval is %d, ReportInterval is %d, LogFile is %s \n", conf.address, conf.pollInterval, conf.reportInterval, conf.logfile)
	return nil
}
