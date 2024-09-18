package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// MemStorage Хранилище для метрик. Разные map-ы для разных типов метрик
type MemStorage struct {
	gaugeMap   map[string]float64
	counterMap map[string]int64
}

// Создание хранилища
var store = MemStorage{gaugeMap: make(map[string]float64), counterMap: make(map[string]int64)}

// Константа для кодирования смысла полей после парсинга URL на основе их порядкового номера
// Пример: localhost:8080/update/gauge/metric2/777.4
const (
	update  = 0
	gauge   = 1
	counter = 1
	metric  = 2
	value   = 3
)

func urlToMap(url string) ([]string, error) {
	// Формируем слайс компонентов URL, предварительно отрезав крайние "/"
	splittedUrl := strings.Split(strings.Trim(url, "/"), "/")
	fmt.Println("Trim in action", strings.Trim(url, "/"))
	// Если длина разобранного URL не больше 2-х -- недостаток указания метрики/значения, возвращаем StatusNotFound
	if len(splittedUrl) <= 3 {
		fmt.Println("ERROR: URL is too short")
		return splittedUrl, errors.New("URL is too short")
	}
	// Если длина разобранного URL больше 4 -- в URL что-то лишнее
	if len(splittedUrl) > 4 {
		fmt.Println("ERROR: URL is too long")
		return splittedUrl, errors.New("URL is too long")
	}
	fmt.Println("urlToMap:", splittedUrl)
	return splittedUrl, nil
}

func gaugeMetricHandler(w http.ResponseWriter, r *http.Request) {
	splittedUrl, err := urlToMap(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Println("URL.String is:", r.URL.String(), "\nsplittedUrl is:", splittedUrl, "len is:", len(splittedUrl))
	if s, err := strconv.ParseFloat(splittedUrl[value], 64); err == nil {
		// записываем метрику в хранилище
		fmt.Println("metric is:", splittedUrl[metric], "\nvalue is:", splittedUrl[value])
		store.gaugeMap[splittedUrl[metric]] = s
		// Формируем ответ
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Println("Response is:", w)
	} else {
		fmt.Println("ERROR: There is no metric or wrong metric type -- must be float64")
		w.WriteHeader(http.StatusBadRequest)
	}
	fmt.Println(store)
}

func counterMetricHandler(w http.ResponseWriter, r *http.Request) {
	splittedUrl, err := urlToMap(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Println("URL.String is:", r.URL.String(), "\nsplittedUrl is:", splittedUrl, "len is:", len(splittedUrl))
	if s, err := strconv.ParseInt(splittedUrl[value], 10, 64); err == nil {
		// записываем метрику в хранилище
		fmt.Println("metric is:", splittedUrl[metric], "\nvalue is:", splittedUrl[value])
		if _, ok := store.counterMap[splittedUrl[metric]]; ok {
			store.counterMap[splittedUrl[metric]] += s
		} else {
			store.counterMap[splittedUrl[metric]] = s
		}
		// Формируем ответ
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Println("Response is:", w)
	} else {
		fmt.Println(s, "There is no metric or wrong metric type -- must be int")
		w.WriteHeader(http.StatusBadRequest)
	}
	fmt.Println(store)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/gauge/", gaugeMetricHandler)
	mux.HandleFunc("/update/counter/", counterMetricHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
