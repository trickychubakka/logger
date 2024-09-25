package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"server/storage/memstorage"
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
		return splittedURL, errors.New("URL is too long")
	}
	fmt.Println("urlToMap:", splittedURL)
	return splittedURL, nil
}

// metricHandler handler обработки запросов
func MetricHandler(w http.ResponseWriter, r *http.Request) {
	splittedURL, err := urlToMap(r.URL.String())
	if err != nil {
		fmt.Println("ERROR:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// metricHandler Обработка gauge метрики
	if splittedURL[metricType] == "gauge" {
		if val, err := strconv.ParseFloat(splittedURL[metricValue], 64); err == nil {
			if err := store.UpdateGauge(splittedURL[metricName], val); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			fmt.Println("ERROR: There is no metric or wrong metric value type -- must be float64")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// metricHandler Обработка counter метрик
	} else if splittedURL[metricType] == "counter" {
		if val, err := strconv.ParseInt(splittedURL[metricValue], 10, 64); err == nil {
			if err := store.UpdateCounter(splittedURL[metricName], val); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			fmt.Println("ERROR: There is no metric or wrong metric value type -- must be int64")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Неправильный тип метрики
	} else {
		fmt.Println("ERROR: Wrong metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Формируем ответ
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Println(store)
}
