// Package memstorage -- пакет с реализацией In Memory типа хранилища.
package memstorage

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"logger/internal/storage"
)

// MemStorage inmemory хранилище для метрик.
type MemStorage struct {
	gaugeMap   map[string]float64 // Map метрик типа gauge.
	counterMap map[string]int64   // Map метрик типа counter.
}

// New -- конструктор объекта хранилища MemStorage.
func New(_ context.Context) (MemStorage, error) {
	return MemStorage{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}, nil
}

// UpdateGauge -- реализация метода изменения Gauge метрики.
func (ms MemStorage) UpdateGauge(_ context.Context, key string, value float64) error {
	ms.gaugeMap[key] = value
	return nil
}

// UpdateCounter -- реализация метода изменения Counter метрики.
func (ms MemStorage) UpdateCounter(_ context.Context, key string, value int64) error {
	ms.counterMap[key] += value
	return nil
}

// UpdateBatch -- реализация метода изменения набора метрик, описанного через массив объектов Metrics.
func (ms MemStorage) UpdateBatch(_ context.Context, metrics []storage.Metrics) error {
	log.Println("UpdateBatch. Start Update batch, storage now is :", ms)
	if len(metrics) == 0 {
		log.Println("UpdateBatch. No metrics to update im []Metrics")
		return nil
	}
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			ms.gaugeMap[metric.ID] = *metric.Value
		case "counter":
			log.Println("UpdateBatch: memstorage update counter ", metric.ID, "value, before:", ms.counterMap[metric.ID], "updating with delta :", *metric.Delta)
			ms.counterMap[metric.ID] += *metric.Delta
			log.Println("UpdateBatch: memstorage update counter value, after:", ms.counterMap[metric.ID])
		}
	}
	log.Println("UpdateBatch. End Update batch")
	return nil
}

// GetGauge -- реализация метода получения Gauge метрики по ее названию.
func (ms MemStorage) GetGauge(_ context.Context, key string) (float64, error) {
	val, ok := ms.gaugeMap[key]
	if !ok {
		return 0, errors.New("no value for key " + key)
	}
	return val, nil
}

// GetCounter -- реализация метода получения Counter метрики по ее названию.
func (ms MemStorage) GetCounter(_ context.Context, key string) (int64, error) {
	val, ok := ms.counterMap[key]
	if !ok {
		return 0, errors.New("no value for key " + key)
	}
	return val, nil
}

// GetValue -- реализация метода получения любой метрики по ее типу и названию.
func (ms MemStorage) GetValue(_ context.Context, t string, key string) (any, error) {
	if t == "counter" {
		val, ok := ms.counterMap[key]
		if !ok {
			return nil, errors.New("no value for key " + key)
		}
		return val, nil
	} else if t == "gauge" {
		val, ok := ms.gaugeMap[key]
		if !ok {
			return nil, errors.New("no value for key " + key)
		}
		return val, nil
	} else {
		return nil, errors.New("wrong metric type")
	}
}

// GetAllGaugesMap -- реализация метода получения всех метрик типа Gauge.
func (ms MemStorage) GetAllGaugesMap(_ context.Context) (map[string]float64, error) {
	return ms.gaugeMap, nil
}

// GetAllCountersMap -- реализация метода получения всех метрик типа Counter.
func (ms MemStorage) GetAllCountersMap(_ context.Context) (map[string]int64, error) {
	return ms.counterMap, nil
}

// GetAllMetrics -- реализация метода получения всех метрик.
func (ms MemStorage) GetAllMetrics(_ context.Context) (any, error) {
	return ms, nil
}

// Close -- заглушка метода закрытия соединения.
func (ms MemStorage) Close() error {
	return nil
}

// Структура для использования в Marshal и Unmarshal функциях.
type tmpMemStorage struct {
	GaugeMap   map[string]float64
	CounterMap map[string]int64
}

// Unmarshal функция для Unmarshal private полей структуры MemStorage.
func Unmarshal(data []byte, stor *MemStorage) error {
	tmp := tmpMemStorage{
		GaugeMap:   make(map[string]float64),
		CounterMap: make(map[string]int64),
	}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	stor.gaugeMap = tmp.GaugeMap
	stor.counterMap = tmp.CounterMap
	return nil
}

// Marshal функция для Marshal private полей структуры MemStorage.
func Marshal(stor any) ([]byte, error) {
	tmp := tmpMemStorage{
		GaugeMap:   make(map[string]float64),
		CounterMap: make(map[string]int64),
	}
	stor = stor.(MemStorage)
	tmp.GaugeMap = stor.(MemStorage).gaugeMap
	tmp.CounterMap = stor.(MemStorage).counterMap
	return json.Marshal(tmp)
}
