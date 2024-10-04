package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

type Config struct {
	pollInterval   int
	reportInterval int
	handlerURL     string
}

// initConfig функция инициализации конфигурации агента с использованием параметров командной строки
func initConfig(h, r, p string, conf *Config) error {
	if _, err := url.ParseRequestURI(h); err != nil {
		return err
	}
	conf.handlerURL = h

	if c, err := strconv.Atoi(r); err == nil {
		conf.reportInterval = c
	} else {
		return err
	}

	if c, err := strconv.Atoi(p); err == nil {
		conf.pollInterval = c
	} else {
		return err
	}

	if conf.pollInterval > conf.reportInterval {
		return errors.New("poll interval must be less than report interval")
	}
	return nil
}

type Metrics struct {
	gaugeMap   map[string]float64
	counterMap map[string]int64
}

func NewMetricsObj() Metrics {
	return Metrics{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}
}

// MetricsPolling -- заполнение словаря метрик перебором всех полей структуры MemStats через reflect
// с выбором метрик небходимых типов
func MetricsPolling(metrics *Metrics) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	val := reflect.ValueOf(memStats)
	numfield := reflect.ValueOf(memStats).NumField()

	for x := 0; x < numfield; x++ {
		//fmt.Printf("Name field: `%s`  Type: `%s`\n", reflect.TypeOf(memStats).Elem().Field(x).Name,
		field := val.Field(x)
		//fmt.Printf("\tField %v: %v - val :%v\n", val.Type().Field(x).Name, field.Type(), field)
		switch field.Kind() {
		case reflect.Uint64:
			metrics.gaugeMap[val.Type().Field(x).Name] = float64(field.Uint())
		case reflect.Uint32:
			metrics.gaugeMap[val.Type().Field(x).Name] = float64(field.Uint())
		case reflect.Float64:
			metrics.gaugeMap[val.Type().Field(x).Name] = field.Float()
		default:
			//fmt.Printf("Unsupported type: %v\n", field.Kind())
		}
	}
	//fmt.Println(metrics.gaugeMap, "\n", len(metrics.gaugeMap))
	metrics.counterMap["PollCount"] = 1
	metrics.gaugeMap["RandomValue"] = rand.Float64()

	return nil
}

// SendRequest выполнение запроса с метриками
func SendRequest(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "text/plain")
	fmt.Println("req.Header is:", req.Header)

	// Отсылка сформированного запроса req. Если сервер не отвечает -- работа агента завершается
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println("Header is:", response.Header)
	//io.Copy(os.Stdout, response.Body) // Вывод ответа в консоль
	err = response.Body.Close()
	if err != nil {
		return response, err
	}

	return response, nil
}

// SendMetrics отсылка метрик на сервер
func SendMetrics(metrics *Metrics, c string) error {
	client := &http.Client{}

	// Цикл для отсылки метрик типа gaugeMap
	for m := range metrics.gaugeMap {
		reqURL := c + "/gauge/" + m + "/" + fmt.Sprintf("%v", metrics.gaugeMap[m])
		fmt.Println(m, "=>", metrics.gaugeMap[m], "url:", reqURL)

		response, err := SendRequest(client, reqURL)
		if err != nil {
			return err
		}
		response.Body.Close()
	}

	// Цикл для отсылки метрик типа counterMap
	for m := range metrics.counterMap {
		reqURL := c + "/counter/" + m + "/" + fmt.Sprintf("%v", metrics.counterMap[m])
		fmt.Println(m, "=>", metrics.counterMap[m], "url:", reqURL)

		response, err := SendRequest(client, reqURL)
		if err != nil {
			return err
		}
		response.Body.Close()
	}
	return nil
}

func main() {

	parseFlags()

	conf := new(Config)
	if err := initConfig(AddressFlag, ReportIntervalFlag, PollingIntervalFlag, conf); err != nil {
		panic(err)
	}
	fmt.Printf("Address is %s, PollInterval is %d, ReportInterval is %d", conf.handlerURL, conf.pollInterval, conf.reportInterval)

	metrics := NewMetricsObj()
	for {
		for i := 0; i < conf.reportInterval; i = i + conf.pollInterval {
			if err := MetricsPolling(&metrics); err != nil {
				fmt.Println(err)
			}
			fmt.Println("\nmetrics:", metrics)
			time.Sleep(time.Duration(conf.pollInterval) * time.Second)
		}
		if err := SendMetrics(&metrics, "http://"+conf.handlerURL+"/update"); err != nil {
			fmt.Println(err)
		}
	}
}
