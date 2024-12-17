package main

import (
	"fmt"
	"log"
	"logger/internal"
	"os"
)

// FlagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
var FlagTest = false

func main() {

	if err := initConfig(&config); err != nil {
		log.Println("AGENT Panic from initConfig", err)
		panic(err)
	}

	if config.Logfile != "" {
		file, err := os.OpenFile(config.Logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file:", err)
		}
		log.SetOutput(file)
		defer file.Close()
	}

	fmt.Printf("\nAddress is %s, PollInterval is %d, ReportInterval is %d, LogFile is %s \n", config.Address, config.PollInterval, config.ReportInterval, config.Logfile)

	myMetrics := internal.NewMetricsStorageObj()

	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("%v", p)
			log.Println("Panic recovering -> main:", err)
			log.Println("recovered from panic in main")
		}
		log.Println("Start run after recovering")
		run(myMetrics, &config)
	}()

	log.Println("Start run")
	run(myMetrics, &config)
}
