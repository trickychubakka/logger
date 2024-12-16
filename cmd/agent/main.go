package main

import (
	"context"
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
func metricsPolling(ctx context.Context, m *sync.RWMutex, myMetrics *internal.MetricsStorage, config *conf.AgentConfig) error {
	log.Println("start metricsPolling goroutine")
	counter := 1
	for {
		select {
		case <-ctx.Done():
			log.Println("STOP metricsPolling goroutine")
			return nil
		default:
			if counter == config.PollInterval {
				log.Println("metricsPolling goroutine polling")
				m.Lock()
				if err := internal.MetricsPolling(myMetrics); err != nil {
					log.Println("error in metricsPolling :", err)
					return err
				}
				m.Unlock()
				counter = 0
			}
			time.Sleep(1 * time.Second)
			counter++
		}
	}
}

// gopsMetricsPolling функция сбора метрик, собранных через gopsutil
func gopsMetricsPolling(ctx context.Context, m *sync.RWMutex, myMetrics *internal.MetricsStorage, config *conf.AgentConfig) error {
	log.Println("start gopsMetricsPolling goroutine")
	counter := 1
	for {
		select {
		case <-ctx.Done():
			log.Println("STOP gopsMetricsPolling goroutine")
			return nil
		default:
			if counter == config.PollInterval {
				log.Println("gopsMetricsPolling goroutine polling")
				m.Lock()
				if err := internal.GopsMetricPolling(myMetrics); err != nil {
					log.Println("error in metricsPolling :", err)
					return err
				}
				m.Unlock()
				counter = 0
			}
			time.Sleep(1 * time.Second)
			counter++
		}
	}
}

// metricReport функция отсылки метрик на сервер
func metricsReport(ctx context.Context, m *sync.RWMutex, myMetrics *internal.MetricsStorage, config *conf.AgentConfig) error {
	log.Println("start metricsReport goroutine")
	counter := 1
	for {
		select {
		case <-ctx.Done():
			log.Println("STOP metricsReport goroutine")
			return nil
		default:
			if counter == config.ReportInterval {
				//time.Sleep(time.Duration(config.ReportInterval) * time.Second)
				log.Println("run. SendMetricsJSONBatch start. myMetrics is:", myMetrics)
				m.RLock()
				if err := internal.SendMetricsJSONBatch(myMetrics, "http://"+config.Address+"/updates", config); err != nil {
					log.Println("main: error from SendMetricsJSONBatch:", err)
					return err
				}
				m.RUnlock()
				counter = 0
			}
			time.Sleep(1 * time.Second)
			counter++
		}
	}
}

// Run функция запуска горутин polling-а метрик и их отсылки на сервер
func run(myMetrics internal.MetricsStorage, config *conf.AgentConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	var m sync.RWMutex
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := metricsPolling(ctx, &m, &myMetrics, config)
		if err != nil {
			log.Panicf("metricsPolling error %s", errors.Unwrap(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := gopsMetricsPolling(ctx, &m, &myMetrics, config)
		if err != nil {
			log.Panicf("gopsMetricsPolling error %s", errors.Unwrap(err))
		}
	}()

	log.Println("start metricsReport")
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := metricsReport(ctx, &m, &myMetrics, config)
		if err != nil {
			log.Panicf("metricsReport error %s", errors.Unwrap(err))
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit
	cancel()
	wg.Wait()
	log.Println("Main done")
	log.Println("AGENT STOPPED.")
	os.Exit(1)
}

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
