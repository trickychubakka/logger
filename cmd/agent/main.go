// Logger агент сбора метрик операционной системы и отправки их на logger-сервер
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"logger/conf"
	"logger/internal"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Переменные для вывода информации при старте приложения.
var buildVersion, buildDate, buildCommit string

// FlagTest флаг режима тестирования для отключения парсинга командной строки при тестировании.
var FlagTest = false

const (
	clientDoErrors int = 3       // Максимально допустимое количество ошибок подключения к серверу client.Do error
	addr               = ":6060" // For pprof HTTP server
)

var srv *http.Server

// metricsPolling функция сбора метрик.
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

// gopsMetricsPolling функция сбора метрик, собранных через gopsutil.
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
				//log.Println("gopsMetricsPolling goroutine polling")
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

// metricReport функция отсылки метрик на logger сервер.
func metricsReport(ctx context.Context, m *sync.RWMutex, myMetrics *internal.MetricsStorage, config *conf.AgentConfig) error {
	log.Println("start metricsReport goroutine")
	counter := 1
	errorCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("STOP metricsReport goroutine")
			return nil
		default:
			if counter == config.ReportInterval {
				m.RLock()
				log.Println("run. SendMetricsJSONBatch start. myMetrics is:", myMetrics)
				if err := internal.SendMetricsJSONBatch(myMetrics, "http://"+config.Address+"/updates", config); err != nil {
					// Если это ошибка подключения к серверу client.Do error -- игнорируем clientDoErrors ошибок, после возвращаем err.
					// Если количество ошибок подключения к серверу >= clientDoErrors -- увеличиваем счетчик ошибок errorCount.
					// Если это не client.Do ошибка -- сразу возвращаем error.
					if strings.Contains(err.Error(), "client.Do error") && errorCount >= clientDoErrors {
						log.Println("main: client.Do error from SendMetricsJSONBatch:", err, "errorCount > 3, raise panic")
						return err
					}
					if !strings.Contains(err.Error(), "client.Do error") {
						log.Println("metricsReport, error from SendMetricsJSONBatch:", err)
						return err
					}
					log.Println("metricsReport, client.Do error from SendMetricsJSONBatch:", err, "errorCount is", errorCount, " ignore this error")
					errorCount++
				}
				m.RUnlock()
				counter = 0
			}
			time.Sleep(1 * time.Second)
			counter++
		}
	}
}

// startHTTPServer -- start HTTP server for pprof.
func startHTTPServer(wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{Addr: addr}

	go func() {
		defer wg.Done()
		// always returns error. ErrServerClosed on graceful close.
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
			log.Fatalf("startHTTPServer ListenAndServe(): %v", err)
		}
	}()
	// returning reference so caller can call Shutdown().
	return srv
}

// Run функция запуска горутин polling-а метрик и их отсылки на logger сервер.
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

	// Starting pprof http.server.
	if config.PProfHTTPEnabled {
		log.Println("start pprof web server")
		wg.Add(1)
		srv = startHTTPServer(&wg)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit
	// Graceful shutdown pprof http server if option -t enabled.
	if config.PProfHTTPEnabled {
		if err := srv.Shutdown(ctx); err != nil {
			log.Println("run, failure/timeout shutting down the server gracefully, error is:", err)
		} else {
			log.Println("run, pprof http.server shutdown gracefully")
		}
	}
	cancel()
	wg.Wait()
	log.Println("Main done")
	log.Println("AGENT STOPPED.")
	//os.Exit(1)
	log.Fatal()
}

func main() {

	internal.PrintStartMessage(buildVersion, buildDate, buildCommit)

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
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Println("Failed to close log file:", err)
			}
		}(file)
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
