package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
	handlerURL     = "http://localhost:8080/update"
)

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
		url := c + "/gauge/" + m + "/" + fmt.Sprintf("%v", metrics.gaugeMap[m])
		fmt.Println(m, "=>", metrics.gaugeMap[m], "url:", url)

		if _, err := SendRequest(client, url); err != nil {
			return err
		}
	}

	// Цикл для отсылки метрик типа counterMap
	for m := range metrics.counterMap {
		url := c + "/counter/" + m + "/" + fmt.Sprintf("%v", metrics.counterMap[m])
		fmt.Println(m, "=>", metrics.counterMap[m], "url:", url)

		if _, err := SendRequest(client, url); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	metrics := NewMetricsObj()
	for {
		for i := 0; i < reportInterval; i = i + pollInterval {
			if err := MetricsPolling(&metrics); err != nil {
				fmt.Println(err)
			}
			fmt.Println("\nmetrics:", metrics)
			time.Sleep(pollInterval * time.Second)
		}
		if err := SendMetrics(&metrics, handlerURL); err != nil {
			fmt.Println(err)
		}
	}
}
