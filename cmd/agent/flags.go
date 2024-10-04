package main

import "flag"

var (
	ReportIntervalFlag  string
	PollingIntervalFlag string
	AddressFlag         string
)

// parseFlags функция обработки параметров командной строки
func parseFlags() {
	flag.StringVar(&AddressFlag, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&ReportIntervalFlag, "r", "10", "agent report interval")
	flag.StringVar(&PollingIntervalFlag, "p", "2", "agent poll interval")

	flag.Parse()
}
