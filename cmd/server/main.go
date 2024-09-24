package main

import (
	"errors"
	"fmt"
	"memstorage"
	//"github.com/trickychubakka/logger/cmd/server/storage/memstorage"
	"net/http"
	"strconv"
	"strings"
)

//// MemStorage Хранилище для метрик. Разные map-ы для разных типов метрик
//type MemStorage struct {
//	gaugeMap   map[string]float64
//	counterMap map[string]int64
//}
//
//// Создание хранилища
//var store = MemStorage{gaugeMap: make(map[string]float64), counterMap: make(map[string]int64)}

var store = memstorage.New()

// Константа для кодирования смысла полей после парсинга URL на основе их порядкового номера
// Пример: localhost:8080/update/gauge/metric2/777.4
const (
	metricType  = 1
	metricName  = 2
	metricValue = 3
)

func urlToMap(url string) ([]string, error) {
	// Формируем слайс компонентов URL, предварительно отрезав крайние "/"
	splittedURL := strings.Split(strings.Trim(url, "/"), "/")
	fmt.Println("Trim in action", strings.Trim(url, "/"))
	// Если длина разобранного URL не больше 2-х -- недостаток указания метрики/значения, возвращаем StatusNotFound
	if len(splittedURL) <= 3 {
		fmt.Println("ERROR: URL is too short")
		return splittedURL, errors.New("URL is too short")
	}
	// Если длина разобранного URL больше 4 -- в URL что-то лишнее
	if len(splittedURL) > 4 {
		fmt.Println("ERROR: URL is too long")
		return splittedURL, errors.New("URL is too long")
	}
	fmt.Println("urlToMap:", splittedURL)
	return splittedURL, nil
}

func metricHandler(w http.ResponseWriter, r *http.Request) {
	splittedURL, err := urlToMap(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Обработка gauge метрики
	if splittedURL[metricType] == "gauge" {
		if val, err := strconv.ParseFloat(splittedURL[metricValue], 64); err == nil {
			// записываем метрику в хранилище
			//fmt.Println("metricType is:", splittedURL[metricType], "| metricName is:", splittedURL[metricName], "| metricValue is:", splittedURL[metricValue])
			//store.gaugeMap[splittedURL[metricName]] = val
			if err := store.UpdateGauge(splittedURL[metricName], val); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			fmt.Println("ERROR: There is no metric or wrong metric value type -- must be float64")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Обработка counter метрик
	} else if splittedURL[metricType] == "counter" {
		if val, err := strconv.ParseInt(splittedURL[metricValue], 10, 64); err == nil {
			// записываем метрику в хранилище
			//fmt.Println("metricType is:", splittedURL[metricType], "| metricName is:", splittedURL[metricName], "| metricValue is:", splittedURL[metricValue])
			//store.counterMap[splittedURL[metricName]] += val
			//store.UpdateCounter(splittedURL[metricName], val)
			if err := store.UpdateCounter(splittedURL[metricName], val); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			fmt.Println("There is no metric or wrong metric value type -- must be int")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Неправильный тип метрики
	} else {
		fmt.Println("Wrong metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Формируем ответ
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//fmt.Println("Response is:", w)
	fmt.Println(store)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", metricHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
