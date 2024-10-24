package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"logger/internal/storage/memstorage"
	"net/http"
	"strconv"
	"strings"
)

var store = memstorage.New()

// Константа для кодирования смысла полей после парсинга URL на основе их порядкового номера
// Пример: localhost:8080/update/gauge/metric2/777.4
const (
	metricType  = 1
	metricName  = 2
	metricValue = 3
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// urlToMap парсинг URL в map по разделителям "/" с предварительным удалением крайних "/"
func urlToMap(url string) ([]string, error) {
	splittedURL := strings.Split(strings.Trim(url, "/"), "/")
	// Если длина разобранного URL не больше 2-х -- недостаток указания метрики/значения, возвращаем StatusNotFound
	if len(splittedURL) <= 3 {
		return splittedURL, errors.New("URL is too short")
	}
	// Если длина разобранного URL больше 4 -- в URL что-то лишнее
	if len(splittedURL) > 4 {
		log.Println("Error in urlToMap: URL is too long")
		return splittedURL, errors.New("URL is too long")
	}
	//fmt.Println("urlToMap:", splittedURL)
	return splittedURL, nil
}

//func urlToMetric(url string) (Metric, error) {}

// MetricsHandler -- Gin handler обработки запросов по изменениям метрик через URL
func MetricsHandler(c *gin.Context) {
	//fmt.Println("Request Header is:", c.Header)
	splittedURL, err := urlToMap(c.Request.URL.String())
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	// metricHandler Обработка gauge метрики
	if splittedURL[metricType] == "gauge" {
		if val, err := strconv.ParseFloat(splittedURL[metricValue], 64); err == nil {
			if err := store.UpdateGauge(splittedURL[metricName], val); err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
		} else {
			log.Println("Error in MetricHandler: There is no metric or wrong metric value type -- must be float64")
			c.Status(http.StatusBadRequest)
			return
		}
		// metricHandler Обработка counter метрик
	} else if splittedURL[metricType] == "counter" {
		if val, err := strconv.ParseInt(splittedURL[metricValue], 10, 64); err == nil {
			if err := store.UpdateCounter(splittedURL[metricName], val); err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
		} else {
			log.Println("Error in MetricHandler: There is no metric or wrong metric value type -- must be int64")
			c.Status(http.StatusBadRequest)
			return
		}
		// Неправильный тип метрики
	} else {
		log.Println("Error in MetricHandler: Wrong metric type")
		c.Status(http.StatusBadRequest)
		return
	}
	log.Println("Requested PLAIN metric UPDATE with next metric")
	// Формируем ответ
	//c.Header("content-type", "text/plain; charset=utf-8")
	c.Header("content-type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	//fmt.Println(store)
}

// MetricHandlerJSON -- Gin handler обработки запросов по изменениям метрик через JSON в Body
func MetricHandlerJSON(c *gin.Context) {
	log.Println("We are in MetricHandlerJSON")
	jsn, err := io.ReadAll(c.Request.Body)
	if err != nil {
		http.Error(c.Writer, "Error in json body read", http.StatusInternalServerError)
		return
	}

	var tmpMetric Metrics

	err = json.Unmarshal(jsn, &tmpMetric)
	if err != nil {
		log.Println("Error in json body read 2", err, "jsn is:", string(jsn))
		c.Status(http.StatusBadRequest)
		return
	}

	// TODO логирование запроса
	log.Println("Requested JSON metric UPDATE with next metric", tmpMetric)

	if tmpMetric.MType == "gauge" {
		if err := store.UpdateGauge(tmpMetric.ID, *tmpMetric.Value); err != nil {
			log.Println("Error in UpdateGauge:", err)
			c.Status(http.StatusInternalServerError)
			return
		}
	} else if tmpMetric.MType == "counter" {
		if err := store.UpdateCounter(tmpMetric.ID, *tmpMetric.Delta); err != nil {
			log.Println("Error in UpdateCounter:", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		// обновляем во временном объекте метрики значение Counter-а для выдачи его в response
		if *tmpMetric.Delta, err = store.GetCounter(tmpMetric.ID); err != nil {
			log.Println("Error in GetCounter:", err)
			c.Status(http.StatusInternalServerError)
			return
		}
	} else {
		log.Println("Error in MetricHandlerJSON: Wrong metric type")
		c.Status(http.StatusBadRequest)
		return
	}

	j2 := io.NopCloser(bytes.NewBuffer(jsn))
	log.Println("Request from j2:", j2)

	resp, err := json.Marshal(tmpMetric)
	log.Println("MetricHandler after json.Marshal, response is ", resp)
	log.Println("MetricHandler after json.Marshal, string response is ", string(resp))
	if err != nil {
		log.Println("Error in json.Marshal in handlers:", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("content-type", "application/json")
	c.Status(http.StatusOK)
	c.Writer.Write(resp)
	//c.Writer.WriteString(string(resp))
}

// GetAllMetrics получить все метрики
func GetAllMetrics(c *gin.Context) {
	// Get all Gauge metrics
	if metrics, err := store.GetAllGaugesMap(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	} else {
		c.Header("content-type", "text/html; charset=utf-8")
		c.String(http.StatusOK, "Gauge metrics:")
		c.IndentedJSON(http.StatusOK, metrics)
	}
	// Get all Counter metrics
	if metrics, err := store.GetAllCountersMap(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	} else {
		c.Header("content-type", "text/html; charset=utf-8")
		c.String(http.StatusOK, "\nCounter metrics:")
		c.IndentedJSON(http.StatusOK, metrics)
	}
	//c.Header("content-type", "text/plain; charset=utf-8")
}

// GetMetric получить значение метрики
func GetMetric(c *gin.Context) {
	splittedURL, err := urlToMap(c.Request.URL.String())
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
	val, err := store.GetValue(splittedURL[metricType], splittedURL[metricName])
	if err != nil {
		fmt.Println("Error in GetMetric:", err)
		c.Status(http.StatusNotFound)
	} else {
		switch v := val.(type) {
		case float64:
			{
				c.String(http.StatusOK, fmt.Sprintf("%g", v))
			}
		case int64:
			c.String(http.StatusOK, fmt.Sprintf("%d", v))
		}
	}
}

// GetMetricJSON получить значение метрики через JSON
func GetMetricJSON(c *gin.Context) {
	jsn, err := io.ReadAll(c.Request.Body)
	log.Println("GetMetricJSON, jsn after ReadAll:", string(jsn))
	if err != nil {
		http.Error(c.Writer, "Error in json body read", http.StatusInternalServerError)
		return
	}

	var tmpMetric Metrics

	err = json.Unmarshal(jsn, &tmpMetric)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if tmpMetric.MType == "gauge" {
		var val float64
		val, err = store.GetGauge(tmpMetric.ID)
		tmpMetric.Value = &val
	}
	if tmpMetric.MType == "counter" {
		var delta int64
		delta, err = store.GetCounter(tmpMetric.ID)
		tmpMetric.Delta = &delta
	}
	if err != nil {
		log.Println("Requested metric value with status 404", tmpMetric)
		//log.Println("error is", err)
		c.Status(http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(tmpMetric)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	log.Println("Requested metric value with status 200", tmpMetric)
	j2 := io.NopCloser(bytes.NewBuffer(resp))
	log.Println("Request value from j2:", j2)

	c.Header("content-type", "application/json")
	c.Status(http.StatusOK)
	c.Writer.Write(resp)
}
