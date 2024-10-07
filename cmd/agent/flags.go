package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	pollInterval   int
	reportInterval int
	address        string
}

// initConfig функция инициализации конфигурации агента с использованием параметров командной строки
func initConfig(conf *Config) error {

	var (
		ReportIntervalFlag string
		PollIntervalFlag   string
		AddressFlag        string
	)

	if !flagTest {
		flag.StringVar(&AddressFlag, "a", "localhost:8080", "address and port to run server")
		flag.StringVar(&ReportIntervalFlag, "r", "10", "agent report interval")
		flag.StringVar(&PollIntervalFlag, "p", "2", "agent poll interval")

		flag.Parse()
	}
	// address processing
	if envAddressFlag := os.Getenv("ADDRESS"); envAddressFlag != "" {
		fmt.Println("env var ADDRESS was specified, use ADDRESS =", envAddressFlag)
		AddressFlag = envAddressFlag
	}
	if _, err := url.ParseRequestURI(AddressFlag); err != nil {
		return err
	}
	conf.address = AddressFlag

	// reportInterval processing
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		fmt.Println("env var REPORT_INTERVAL was specified, use REPORT_INTERVAL =", envReportInterval)
		ReportIntervalFlag = envReportInterval
	}

	if c, err := strconv.Atoi(ReportIntervalFlag); err == nil {
		conf.reportInterval = c
	} else {
		return err
	}

	// PollInterval processing
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		fmt.Println("env var POLL_INTERVAL was specified, use POLL_INTERVAL =", envPollInterval)
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

	fmt.Printf("Address is %s, PollInterval is %d, ReportInterval is %d", conf.address, conf.pollInterval, conf.reportInterval)
	return nil
}
