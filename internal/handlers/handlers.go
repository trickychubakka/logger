package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
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

// urlToMap парсинг URL в map по разделителям "/" с предварительным удалением крайних "/"
func urlToMap(url string) ([]string, error) {
	splittedURL := strings.Split(strings.Trim(url, "/"), "/")
	// Если длина разобранного URL не больше 2-х -- недостаток указания метрики/значения, возвращаем StatusNotFound
	if len(splittedURL) <= 3 {
		return splittedURL, errors.New("URL is too short")
	}
	// Если длина разобранного URL больше 4 -- в URL что-то лишнее
	if len(splittedURL) > 4 {
		fmt.Println("Error in urlToMap: URL is too long")
		return splittedURL, errors.New("URL is too long")
	}
	//fmt.Println("urlToMap:", splittedURL)
	return splittedURL, nil
}

//// MetricHandler handler обработки запросов
//func MetricHandler(w http.ResponseWriter, r *http.Request) {
//	fmt.Println("Request Header is:", r.Header)
//	splittedURL, err := urlToMap(r.URL.String())
//	if err != nil {
//		fmt.Println("ERROR:", err)
//		w.WriteHeader(http.StatusNotFound)
//		return
//	}
//	// metricHandler Обработка gauge метрики
//	if splittedURL[metricType] == "gauge" {
//		if val, err := strconv.ParseFloat(splittedURL[metricValue], 64); err == nil {
//			if err := store.UpdateGauge(splittedURL[metricName], val); err != nil {
//				w.WriteHeader(http.StatusInternalServerError)
//				return
//			}
//		} else {
//			fmt.Println("ERROR: There is no metric or wrong metric value type -- must be float64")
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//		// metricHandler Обработка counter метрик
//	} else if splittedURL[metricType] == "counter" {
//		if val, err := strconv.ParseInt(splittedURL[metricValue], 10, 64); err == nil {
//			if err := store.UpdateCounter(splittedURL[metricName], val); err != nil {
//				w.WriteHeader(http.StatusInternalServerError)
//				return
//			}
//		} else {
//			fmt.Println("ERROR: There is no metric or wrong metric value type -- must be int64")
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//		// Неправильный тип метрики
//	} else {
//		fmt.Println("ERROR: Wrong metric type")
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//	// Формируем ответ
//	w.Header().Set("content-type", "text/plain; charset=utf-8")
//	w.WriteHeader(http.StatusOK)
//	fmt.Println(store)
//}

// MetricsHandler -- Gin handler обработки запросов
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
			fmt.Println("Error in MetricHandler: There is no metric or wrong metric value type -- must be float64")
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
			fmt.Println("Error in MetricHandler: There is no metric or wrong metric value type -- must be int64")
			c.Status(http.StatusBadRequest)
			return
		}
		// Неправильный тип метрики
	} else {
		fmt.Println("Error in MetricHandler: Wrong metric type")
		c.Status(http.StatusBadRequest)
		return
	}
	// Формируем ответ
	c.Header("content-type", "text/plain; charset=utf-8")
	c.Status(http.StatusOK)
	//fmt.Println(store)
}

// GetAllMetrics получить все метрики
func GetAllMetrics(c *gin.Context) {
	// Get all Gauge metrics
	if metrics, err := store.GetAllGaugesMap(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	} else {
		c.String(http.StatusOK, "Gauge metrics:")
		c.IndentedJSON(http.StatusOK, metrics)
	}
	// Get all Counter metrics
	if metrics, err := store.GetAllCountersMap(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	} else {
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
				//fmt.Println(fmt.Sprintf("%g", val.(float64)))
				//c.String(http.StatusOK, fmt.Sprintf("%g", val.(float64)))
				c.String(http.StatusOK, fmt.Sprintf("%g", v))
			}
		case int64:
			//c.String(http.StatusOK, fmt.Sprintf("%d", val.(int64)))
			c.String(http.StatusOK, fmt.Sprintf("%d", v))
		}
	}

}
