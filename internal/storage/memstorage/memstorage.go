package memstorage

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"logger/internal/storage"
)

// MemStorage inmemory хранилище для метрик. Разные map-ы для разных типов метрик.
type MemStorage struct {
	gaugeMap   map[string]float64
	counterMap map[string]int64
}

func New(_ context.Context) (MemStorage, error) {
	return MemStorage{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}, nil
}

func (ms MemStorage) UpdateGauge(_ context.Context, key string, value float64) error {
	ms.gaugeMap[key] = value
	return nil
}

func (ms MemStorage) UpdateCounter(_ context.Context, key string, value int64) error {
	ms.counterMap[key] += value
	return nil
}

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

func (ms MemStorage) GetGauge(_ context.Context, key string) (float64, error) {
	val, ok := ms.gaugeMap[key]
	if !ok {
		return 0, errors.New("no value for key " + key)
	}
	return val, nil
}

func (ms MemStorage) GetCounter(_ context.Context, key string) (int64, error) {
	val, ok := ms.counterMap[key]
	if !ok {
		return 0, errors.New("no value for key " + key)
	}
	return val, nil
}

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

func (ms MemStorage) GetAllGaugesMap(_ context.Context) (map[string]float64, error) {
	return ms.gaugeMap, nil
}

func (ms MemStorage) GetAllCountersMap(_ context.Context) (map[string]int64, error) {
	return ms.counterMap, nil
}

func (ms MemStorage) GetAllMetrics(_ context.Context) (any, error) {
	//func (ms MemStorage) GetAllMetrics(_ context.Context) (interface{}, error) {
	return ms, nil
}

func (ms MemStorage) Close() error {
	return nil
}

// Временная структура для использования в Unmarshal методе
type tmpMemStorage struct {
	GaugeMap   map[string]float64
	CounterMap map[string]int64
}

// Unmarshal функция для Unmarshal private полей структуры MemStorage
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

func Marshal(stor any) ([]byte, error) {
	tmp := tmpMemStorage{
		GaugeMap:   make(map[string]float64),
		CounterMap: make(map[string]int64),
	}
	switch stor.(type) {
	case MemStorage:
		tmp.GaugeMap = stor.(MemStorage).gaugeMap
		tmp.CounterMap = stor.(MemStorage).counterMap
	}
	return json.Marshal(tmp)
}
