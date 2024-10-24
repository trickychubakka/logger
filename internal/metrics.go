package internal

import (
	"bytes"
	"compress/gzip"
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

var client = &http.Client{}

//func InitHTTPClient() {
//	tr := &http.Transport{
//		MaxIdleConnsPerHost: 1024,
//		TLSHandshakeTimeout: 0 * time.Second,
//	}
//	client = &http.Client{Transport: tr}
//}

func NewMetricsStorageObj() MetricsStorage {
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
	//func SendRequest(url string, body io.Reader, contentType string) (*http.Response, error) {
	log.Println("body is:", body)
	b, err := io.ReadAll(body)
	log.Println("body io.Reader in start of SendRequest is", b) //, string(b))

	var buf bytes.Buffer
	zb := gzip.NewWriter(&buf)

	_, err = zb.Write(b)
	err = zb.Close()

	bb := bytes.NewReader(buf.Bytes())
	//err = zb.Close() NO!!!
	//
	log.Println("bb before NewRequest is:", bb)

	//req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf.Bytes()))
	req, err := http.NewRequest(http.MethodPost, url, bb)
	//req.Close = true //???
	defer req.Body.Close()

	//req, err := http.NewRequest(http.MethodPost, url, body) // без gzip
	req.Close = true
	if err != nil {
		log.Println("Panic in SendRequest(): ", err)
		panic(err)
	}

	req.Header.Set("Content-Type", contentType)
	// gzip
	req.Header.Set("Content-Encoding", "gzip")

	log.Println("req.Header is:", req.Header)
	//var response *http.Response
	// Отсылка сформированного запроса req. Если сервер не отвечает -- работа агента завершается

	response, err := client.Do(req)
	//io.Copy(io.Discard, response.Body) //
	//if response, err := client.Do(req); response != nil {
	if response != nil {
		log.Println("response is:", response)
		defer response.Body.Close()
	} else if err != nil {
		log.Println("WARNING!!!!!", err, "; response is", response)
		panic(err)
	}

	//defer response.Body.Close()

	//log.Println("Header is:", response.Header)

	return response, nil
}

//
// ###func SendRequest(client *http.Client, url string, body io.Reader, contentType string) (*http.Response, error) {
//func SendRequest(url string, body []byte, contentType string) (*http.Response, error) {
//	//log.Println("body is:", body)
//	//b, err := io.ReadAll(body)
//	//log.Println("body io.Reader in start of SendRequest is", b) //, string(b))
//
//	var buf bytes.Buffer
//	zb := gzip.NewWriter(&buf)
//
//	_, err := zb.Write(body)
//	//err = zb.Close()
//
//	bb := bytes.NewReader(buf.Bytes())
//	err = zb.Close()
//	//
//	log.Println("bb before NewRequest is:", bb, "url is:", url)
//
//	//req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf.Bytes()))
//	req, err := http.NewRequest(http.MethodPost, url, bb)
//	//req.Close = true //???
//	defer req.Body.Close()
//
//	//req, err := http.NewRequest(http.MethodPost, url, body) // без gzip
//	//req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body)) // без gzip
//	//req.Close = true
//	if err != nil {
//		log.Println("Panic in SendRequest(): ", err)
//		panic(err)
//	}
//
//	req.Header.Set("Content-Type", contentType)
//	// gzip
//	req.Header.Set("Content-Encoding", "gzip")
//	req.Header.Set("Accept-Encoding", "gzip")
//
//	log.Println("req.Header is:", req.Header)
//	var response *http.Response
//	// Отсылка сформированного запроса req. Если сервер не отвечает -- работа агента завершается
//
//	response, err = client.Do(req)
//	//io.Copy(io.Discard, response.Body) //
//	//if response, err := client.Do(req); response != nil {
//	if response != nil {
//		log.Println("response is:", response)
//		defer response.Body.Close()
//	} else if err != nil {
//		log.Println("WARNING!!!!!", err, "; response is", response)
//		panic(err)
//	}
//
//	//defer response.Body.Close()
//
//	//log.Println("Header is:", response.Header)
//
//	return response, nil
//}

// SendMetrics отсылка метрик на сервер
func SendMetrics(metrics *MetricsStorage, c string) error {
	//client := &http.Client{}
	count := 0

	// Цикл для отсылки метрик типа gaugeMap
	for m := range metrics.gaugeMap {
		count++
		reqURL := c + "/gauge/" + m + "/" + fmt.Sprintf("%v", metrics.gaugeMap[m])
		log.Println(m, "=>", metrics.gaugeMap[m], "url:", reqURL, "count:", count)

		response, err := SendRequest(client, reqURL, nil, "text/plain")
		//response, err := SendRequest(reqURL, nil, "text/plain")
		if err != nil {
			return err
		}
		defer response.Body.Close()

		//response.Body.Close()
		log.Println("response status:", response.Status)
	}

	// Цикл для отсылки метрик типа counterMap
	for m := range metrics.counterMap {
		count++
		reqURL := c + "/counter/" + m + "/" + fmt.Sprintf("%v", metrics.counterMap[m])
		log.Println(m, "=>", metrics.counterMap[m], "url:", reqURL, "count:", count)

		response, err := SendRequest(client, reqURL, nil, "text/plain")
		//response, err := SendRequest(reqURL, nil, "text/plain")
		if err != nil {
			log.Println("Error Send Metrics in SendRequest call:", err)
			return err
		}
		defer response.Body.Close()

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
	//###client := &http.Client{}
	count := 0

	// Цикл для отсылки метрик типа gaugeMap
	for m := range metrics.gaugeMap {
		count++
		log.Println(m, "=>", metrics.gaugeMap[m], "url:", reqURL, "JSON count:", count)
		valGauge := metrics.gaugeMap[m]
		var tmpMetric = Metrics{m, "gauge", nil, &valGauge}

		payload, err := json.Marshal(tmpMetric)
		log.Println("payload in SendMetrics is:", string(payload))
		if err != nil {
			return err
		}
		response, err := SendRequest(client, reqURL, bytes.NewReader(payload), "application/json")
		//response, err := SendRequest(reqURL, bytes.NewReader(payload), "application/json")
		if err != nil {
			log.Println("Error Send Metrics in SendRequest call:", err)
			return err
		}
		defer response.Body.Close()
	}

	// Цикл для отсылки метрик типа counterMap
	for m := range metrics.counterMap {
		count++
		log.Println(m, "=>", metrics.counterMap[m], "url:", reqURL, "JSON count:", count)
		valCounter := metrics.counterMap[m]
		var tmpMetric = Metrics{m, "counter", &valCounter, nil}

		payload, err := json.Marshal(tmpMetric)
		if err != nil {
			return err
		}

		response, err := SendRequest(client, reqURL, bytes.NewReader(payload), "application/json")
		//response, err := SendRequest(reqURL, bytes.NewReader(payload), "application/json")

		if err != nil {
			log.Println("Error in SendMetricsJSON from SendRequest", err)
			return err
		}
		defer response.Body.Close()
	}
	return nil
}

//func SendMetricsJSON(metrics *MetricsStorage, reqURL string) error {
//	//###client := &http.Client{}
//	count := 0
//
//	// Цикл для отсылки метрик типа gaugeMap
//	for m := range metrics.gaugeMap {
//		count++
//		log.Println(m, "=>", metrics.gaugeMap[m], "url:", reqURL, "JSON count:", count)
//		valGauge := metrics.gaugeMap[m]
//		var tmpMetric = Metrics{m, "gauge", nil, &valGauge}
//
//		payload, err := json.Marshal(tmpMetric)
//		log.Println("payload in SendMetrics is:", string(payload))
//		if err != nil {
//			return err
//		}
//		//###response, err := SendRequest(client, reqURL, bytes.NewReader(payload), "application/json")
//		//response, err := SendRequest(reqURL, bytes.NewReader(payload), "application/json")
//		response, err := SendRequest(reqURL, payload, "application/json")
//		if err != nil {
//			log.Println("Error Send Metrics in SendRequest call:", err)
//			return err
//		}
//		defer response.Body.Close()
//	}
//
//	// Цикл для отсылки метрик типа counterMap
//	for m := range metrics.counterMap {
//		count++
//		log.Println(m, "=>", metrics.counterMap[m], "url:", reqURL, "JSON count:", count)
//		valCounter := metrics.counterMap[m]
//		var tmpMetric = Metrics{m, "counter", &valCounter, nil}
//
//		payload, err := json.Marshal(tmpMetric)
//		if err != nil {
//			return err
//		}
//
//		//###response, err := SendRequest(client, reqURL, bytes.NewReader(payload), "application/json")
//		//response, err := SendRequest(reqURL, bytes.NewReader(payload), "application/json")
//		response, err := SendRequest(reqURL, payload, "application/json")
//
//		if err != nil {
//			log.Println("Error in SendMetricsJSON from SendRequest", err)
//			return err
//		}
//		defer response.Body.Close()
//	}
//	return nil
//}

// PingServer -- функция пинга сервера для решения проблемы metrictests
func PingServer(url string, contentType string) (*http.Response, error) {
	log.Println("PING SERVER with url", url)
	var tmpVar int64
	var tmpMetric = Metrics{"Ping", "counter", &tmpVar, nil}
	payload, err := json.Marshal(tmpMetric)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload)) // без gzip
	if err != nil {
		log.Println("Panic in PingServer(): ", err)
		panic(err)
	}

	req.Header.Set("Content-Type", contentType)

	response, err := client.Do(req)
	log.Println("response in PingServer is:", response)
	defer response.Body.Close()
	return response, err
}
