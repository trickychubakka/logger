package main

import (
	"errors"
	"fmt"
	"log"
	"logger/conf"
	"logger/internal"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// FlagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
var FlagTest = false

// metricsPolling функция сбора метрик
func metricsPolling(m *sync.RWMutex, myMetrics *internal.MetricsStorage, config *conf.AgentConfig) error {
	log.Println("start metricsPolling goroutine")
	for {
		log.Println("metricsPolling goroutine polling")
		m.Lock()
		if err := internal.MetricsPolling(myMetrics); err != nil {
			log.Println("error in metricsPolling :", err)
			return err
		}
		m.Unlock()
		time.Sleep(time.Duration(config.PollInterval) * time.Second)
	}
}

// gopsMetricsPolling функция сбора метрик, собранных через gopsutil
func gopsMetricsPolling(m *sync.RWMutex, myMetrics *internal.MetricsStorage, config *conf.AgentConfig) error {
	log.Println("start gopsMetricsPolling goroutine")
	for {
		log.Println("gopsMetricsPolling goroutine polling")
		m.Lock()
		if err := internal.GopsMetricPolling(myMetrics); err != nil {
			log.Println("error in metricsPolling :", err)
			return err
		}
		m.Unlock()
		time.Sleep(time.Duration(config.PollInterval) * time.Second)
	}
}

// metricReport функция отсылки метрик на сервер
func metricsReport(m *sync.RWMutex, myMetrics *internal.MetricsStorage, config *conf.AgentConfig) error {
	log.Println("start metricsReport goroutine")
	for {
		time.Sleep(time.Duration(config.ReportInterval) * time.Second)
		log.Println("run. SendMetricsJSONBatch start. myMetrics is:", myMetrics)
		m.RLock()
		if err := internal.SendMetricsJSONBatch(myMetrics, "http://"+config.Address+"/updates", config); err != nil {
			log.Println("main: error from SendMetricsJSONBatch:", err)
			log.Panicf("%s", errors.Unwrap(err))
		}
		m.RUnlock()
	}
}

// Run функция запуска горутин polling-а метрик и их отсылки на сервер
func run(myMetrics internal.MetricsStorage, config *conf.AgentConfig) {
	var m sync.RWMutex

	go func() {
		err := metricsPolling(&m, &myMetrics, config)
		if err != nil {
			log.Panicf("metricsPolling error %s", errors.Unwrap(err))
		}
	}()

	go func() {
		err := gopsMetricsPolling(&m, &myMetrics, config)
		if err != nil {
			log.Panicf("gopsMetricsPolling error %s", errors.Unwrap(err))
		}
	}()

	log.Println("start metricsReport")
	go func() {
		err := metricsReport(&m, &myMetrics, config)
		if err != nil {
			log.Panicf("metricsReport error %s", errors.Unwrap(err))
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("AGENT STOPPED.")
	os.Exit(1)

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
		run(myMetrics, &config)
	}()

	run(myMetrics, &config)
}
