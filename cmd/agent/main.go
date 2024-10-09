package main

import (
	"fmt"
	"logger/internal"
	"time"
)

var conf Config

// flagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
var flagTest = false

func main() {

	//if err := initConfig(AddressFlag, ReportIntervalFlag, PollingIntervalFlag, conf); err != nil {
	if err := initConfig(&conf); err != nil {
		panic(err)
	}

	fmt.Printf("\nAddress is %s, PollInterval is %d, ReportInterval is %d", conf.address, conf.pollInterval, conf.reportInterval)

	myMetrics := internal.NewMetricsObj()
	for {
		for i := 0; i < conf.reportInterval; i = i + conf.pollInterval {
			if err := internal.MetricsPolling(&myMetrics); err != nil {
				fmt.Println(err)
			}
			fmt.Println("\nmetrics:", myMetrics)
			time.Sleep(time.Duration(conf.pollInterval) * time.Second)
		}
		if err := internal.SendMetrics(&myMetrics, "http://"+conf.address+"/update"); err != nil {
			fmt.Println(err)
		}
	}
}
