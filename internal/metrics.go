package internal

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"io"
	"log"
	"logger/conf"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

type MetricsStorage struct {
	gaugeMap   map[string]float64
	counterMap map[string]int64
}

// Набор из 3-х таймаутов для повтора операции в случае retriable-ошибки
var timeoutsRetryConst = [3]int{1, 3, 5}

var client = &http.Client{}

func NewMetricsStorageObj() MetricsStorage {
	return MetricsStorage{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}
}

// MetricsPolling -- заполнение словаря метрик перебором всех полей структуры MemStats через reflect
// с выбором метрик необходимых типов
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

func GopsMetricPolling(metrics *MetricsStorage) error {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Println("GopsMetricPolling error in mem.VirtualMemory()")
		return err
	}
	metrics.gaugeMap["TotalMemory"] = float64(v.Total)
	metrics.gaugeMap["FreeMemory"] = float64(v.Free)

	c, err := cpu.Percent(0, false)
	if err != nil {
		log.Println("GopsMetricPolling error in cpu.Percent(0, false)")
		return err
	}
	metrics.gaugeMap["CPUutilization1"] = c[0]

	return nil
}

// hashBody функция подписи body отсылаемого сообщения
// Если ключ задан (не равен "" в AgentConfig) -- возвращаем hash, true
// Если ключ не задан -- возвращаем nil, false
func hashBody(body []byte, config *conf.AgentConfig) ([]byte, bool) {
	if config.Key == "" {
		log.Println("config.Key is empty")
		return nil, false
	}
	h := hmac.New(sha256.New, []byte(config.Key))
	h.Write(body)
	dst := h.Sum(nil)
	fmt.Printf("%x", dst)
	return dst, true
}

func SendRequest(client *http.Client, url string, body io.Reader, contentType string, config *conf.AgentConfig) (*http.Response, error) {

	var hash []byte
	var keyBool bool

	if body != nil {

		b, err := io.ReadAll(body)
		if err != nil {
			log.Println("SendRequest. Error reading body:", err)
			return nil, err
		}

		var buf bytes.Buffer
		zb := gzip.NewWriter(&buf)

		if _, err := zb.Write(b); err != nil {
			log.Println("SendRequest. Error gzip body:", err)
			return nil, err
		}

		err = zb.Close()
		if err != nil {
			log.Println("SendRequest. Error closing compress writer:", err)
		}

		rawBody := buf.Bytes()
		//body = bytes.NewReader(buf.Bytes())
		body = bytes.NewReader(rawBody)
		// Считаем hash256 body ПОСЛЕ gzip-упаковки
		hash, keyBool = hashBody(rawBody, config)
		log.Println("hash is:", hash, "keyBool is :", keyBool)
		if err != nil {
			log.Println("SendRequest. Error hashing body:", err)
		}
		log.Printf("HashSHA256 is : %x", hash)
	}

	req, err := http.NewRequest(http.MethodPost, url, body)

	if err != nil {
		log.Println("SendRequest. Error creating request:", err)
		return nil, fmt.Errorf("%s %v", "SendRequest: http.NewRequest error.", err)
		//panic(err)
	}
	if body != nil {
		defer req.Body.Close()
	}

	req.Close = true

	// Устанавливаем Header key HashSHA256, если key определен
	if keyBool {
		req.Header.Set("HashSHA256", hex.EncodeToString(hash))
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", "compress")

	log.Println("req.Header is:", req.Header)

	// Отсылка сформированного запроса req. Если сервер не отвечает -- работа агента завершается
	response, err := client.Do(req)

	if err != nil {
		for i, t := range timeoutsRetryConst {
			log.Println("SendRequest. Trying to recover after ", t, "seconds, attempt number ", i+1)
			time.Sleep(time.Duration(t) * time.Second)
			response, err = client.Do(req)
			if err != nil {
				log.Println("SendRequest: attempt ", i+1, " error")
				if i == 2 {
					//panic(fmt.Errorf("%s %v", "SendRequest: PANIC in SendRequest.", err))
					return nil, fmt.Errorf("%s %v", "SendRequest: client.Do error.", err)
				}
				continue
			}
			return response, nil
		}
	}
	if response != nil {
		log.Println("SendRequest: response is:", response)
		defer response.Body.Close()
	}
	return response, nil
}

// SendMetrics отсылка метрик на сервер
func SendMetrics(metrics *MetricsStorage, c string, config *conf.AgentConfig) error {
	count := 0

	// Цикл для отсылки метрик типа gaugeMap
	for m := range metrics.gaugeMap {
		count++
		reqURL := c + "/gauge/" + m + "/" + fmt.Sprintf("%v", metrics.gaugeMap[m])
		log.Println(m, "=>", metrics.gaugeMap[m], "url:", reqURL, "count:", count)

		response, err := SendRequest(client, reqURL, nil, "text/plain", config)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		log.Println("response status:", response.Status)
	}

	// Цикл для отсылки метрик типа counterMap
	for m := range metrics.counterMap {
		count++
		reqURL := c + "/counter/" + m + "/" + fmt.Sprintf("%v", metrics.counterMap[m])
		log.Println(m, "=>", metrics.counterMap[m], "url:", reqURL, "count:", count)

		response, err := SendRequest(client, reqURL, nil, "text/plain", config)
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
	ID    string   `json:"id"`              // Имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // Значение метрики в случае передачи gauge
}

func SendMetricsJSON(metrics *MetricsStorage, reqURL string, config *conf.AgentConfig) error {
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
		response, err := SendRequest(client, reqURL, bytes.NewReader(payload), "application/json", config)
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

		response, err := SendRequest(client, reqURL, bytes.NewReader(payload), "application/json", config)

		if err != nil {
			log.Println("Error in SendMetricsJSON from SendRequest", err)
			return err
		}
		defer response.Body.Close()
	}
	return nil
}

func MemstorageToMetrics(store MetricsStorage) ([]Metrics, error) {
	var metrics []Metrics
	var tmpMetric Metrics
	for k, v := range store.gaugeMap {
		tmpMetric.ID = k
		tmpMetric.MType = "gauge"
		tmpMetric.Value = &v
		metrics = append(metrics, tmpMetric)
	}
	for k, v := range store.counterMap {
		tmpMetric.ID = k
		tmpMetric.MType = "counter"
		tmpMetric.Delta = &v
		metrics = append(metrics, tmpMetric)
	}
	log.Println("MetricsToMemstorage: []Metrics :", metrics, " -> store :", store)
	return metrics, nil
}

func SendMetricsJSONBatch(metrics *MetricsStorage, reqURL string, config *conf.AgentConfig) error {
	tmpMetrics, err := MemstorageToMetrics(*metrics)
	if err != nil {
		log.Println("Error in SendMetricsJSONBatch:", err)
		return err
	}
	payload, err := json.Marshal(tmpMetrics)
	if err != nil {
		log.Println("SendMetricsJSONBatch error in json.Marshal: ", err)
	}
	log.Println("payload in SendMetricsJSONBatch is:", string(payload))

	response, err := SendRequest(client, reqURL, bytes.NewReader(payload), "application/json", config)
	if err != nil {
		log.Println("SendMetricsJSONBatch: Error from SendRequest call:", err)
		return err
	}
	defer response.Body.Close()
	return nil
}
