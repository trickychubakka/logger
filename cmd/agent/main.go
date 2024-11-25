package main

import (
	"errors"
	"fmt"
	"log"
	"logger/internal"
	"os"
	"time"
)

// FlagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
var FlagTest = false

// Run функция выполнения цикла polling-а метрик
func run(myMetrics internal.MetricsStorage) {

	for {
		for i := 0; i < config.ReportInterval; i = i + config.PollInterval {
			if err := internal.MetricsPolling(&myMetrics); err != nil {
				log.Println(err)
			}
			log.Println("\nmetrics:", myMetrics)
			time.Sleep(time.Duration(config.PollInterval) * time.Second)
		}

		log.Println("run. SendMetricsJSONBatch start. myMetrics is:", myMetrics)
		if err := internal.SendMetricsJSONBatch(&myMetrics, "http://"+config.Address+"/updates", &config); err != nil {
			log.Println("main: error from SendMetricsJSONBatch:", err)
			log.Panicf("%s", errors.Unwrap(err))
		}
	}
}

func main() {

	if err := initConfig(&config); err != nil {
		log.Println("AGENT Panic from initConfig", err)
		panic(err)
	}

	if config.Logfile != "" {
		file, err := os.OpenFile("agent.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
		run(myMetrics)
	}()

	run(myMetrics)
}
