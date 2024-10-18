package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
)

type MetricsStorage struct {
	gaugeMap   map[string]float64
	counterMap map[string]int64
}

func NewMetricsObj() MetricsStorage {
	return MetricsStorage{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}
}

// MetricsPolling -- заполнение словаря метрик перебором всех полей структуры MemStats через reflect
// с выбором метрик небходимых типов
func MetricsPolling(metrics *MetricsStorage) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	val := reflect.ValueOf(memStats)
	numfield := reflect.ValueOf(memStats).NumField()

	for x := 0; x < numfield; x++ {
		field := val.Field(x)
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
	metrics.counterMap["PollCount"] = 1
	metrics.gaugeMap["RandomValue"] = rand.Float64()

	return nil
}

func SendRequest(client *http.Client, url string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", contentType)
	log.Println("req.Header is:", req.Header)

	var response *http.Response

	// Отсылка сформированного запроса req. Если сервер не отвечает -- работа агента завершается
	response, err = client.Do(req)

	if err != nil {
		log.Println("WARNING!!!!!", err)
		//return response, err
		panic(err)
	}

	defer response.Body.Close()

	//log.Println("Header is:", response.Header)

	return response, nil
}

// SendMetrics отсылка метрик на сервер
func SendMetrics(metrics *MetricsStorage, c string) error {

	client := &http.Client{}

	// Цикл для отсылки метрик типа gaugeMap
	for m := range metrics.gaugeMap {
		reqURL := c + "/gauge/" + m + "/" + fmt.Sprintf("%v", metrics.gaugeMap[m])
		log.Println(m, "=>", metrics.gaugeMap[m], "url:", reqURL)

		response, err := SendRequest(client, reqURL, nil, "text/plain")

		if err != nil {
			return err
		}
		//response.Body.Close()
		log.Println("response status:", response.Status)
	}

	// Цикл для отсылки метрик типа counterMap
	for m := range metrics.counterMap {
		reqURL := c + "/counter/" + m + "/" + fmt.Sprintf("%v", metrics.counterMap[m])
		log.Println(m, "=>", metrics.counterMap[m], "url:", reqURL)

		response, err := SendRequest(client, reqURL, nil, "text/plain")
		if err != nil {
			log.Println("Error Send Metrics in SendRequest call:", err)
			return err
		}
		//
		log.Println("response status:", response.Status)
	}
	return nil
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func SendMetricsJSON(metrics *MetricsStorage, reqURL string) error {
	client := &http.Client{}
	// Цикл для отсылки метрик типа gaugeMap
	for m := range metrics.gaugeMap {
		log.Println(m, "=>", metrics.gaugeMap[m], "url:", reqURL)
		valGauge := metrics.gaugeMap[m]
		var tmpMetric = Metrics{m, "gauge", nil, &valGauge}

		payload, err := json.Marshal(tmpMetric)
		if err != nil {
			return err
		}
		_, err = SendRequest(client, reqURL, bytes.NewReader(payload), "application/json")
	}

	// Цикл для отсылки метрик типа counterMap
	for m := range metrics.counterMap {
		log.Println(m, "=>", metrics.counterMap[m], "url:", reqURL)
		valCounter := metrics.counterMap[m]
		var tmpMetric = Metrics{m, "counter", &valCounter, nil}

		payload, err := json.Marshal(tmpMetric)
		if err != nil {
			return err
		}

		_, err = SendRequest(client, reqURL, bytes.NewReader(payload), "application/json")
		if err != nil {
			log.Println("Error in SendMetricsJSON from SendRequest", err)
			return err
		}
	}
	return nil
}
