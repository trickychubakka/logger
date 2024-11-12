package memstorage

import (
	"context"
	"errors"
	"log"
	"logger/internal/storage"
)

// MemStorage inmemory хранилище для метрик. Разные map-ы для разных типов метрик
type MemStorage struct {
	GaugeMap   map[string]float64
	CounterMap map[string]int64
}

func New(_ context.Context) (MemStorage, error) {
	return MemStorage{
		GaugeMap:   make(map[string]float64),
		CounterMap: make(map[string]int64),
	}, nil
}

func (ms MemStorage) UpdateGauge(_ context.Context, key string, value float64) error {
	ms.GaugeMap[key] = value
	return nil
}

func (ms MemStorage) UpdateCounter(_ context.Context, key string, value int64) error {
	ms.CounterMap[key] += value
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
			ms.GaugeMap[metric.ID] = *metric.Value
		case "counter":
			log.Println("UpdateBatch: memstorage update counter ", metric.ID, "value, before:", ms.CounterMap[metric.ID], "updating with delta :", *metric.Delta)
			ms.CounterMap[metric.ID] += *metric.Delta
			log.Println("UpdateBatch: memstorage update counter value, after:", ms.CounterMap[metric.ID])
		}
	}
	log.Println("UpdateBatch. End Update batch")
	return nil
}

func (ms MemStorage) GetGauge(_ context.Context, key string) (float64, error) {
	val, ok := ms.GaugeMap[key]
	if !ok {
		return 0, errors.New("no value for key " + key)
	}
	return val, nil
}

func (ms MemStorage) GetCounter(_ context.Context, key string) (int64, error) {
	val, ok := ms.CounterMap[key]
	if !ok {
		return 0, errors.New("no value for key " + key)
	}
	return val, nil
}

func (ms MemStorage) GetValue(_ context.Context, t string, key string) (any, error) {
	if t == "counter" {
		val, ok := ms.CounterMap[key]
		if !ok {
			return nil, errors.New("no value for key " + key)
		}
		return val, nil
	} else if t == "gauge" {
		val, ok := ms.GaugeMap[key]
		if !ok {
			return nil, errors.New("no value for key " + key)
		}
		return val, nil
	} else {
		return nil, errors.New("wrong metric type")
	}
}

func (ms MemStorage) GetAllGaugesMap(_ context.Context) (map[string]float64, error) {
	return ms.GaugeMap, nil
}

func (ms MemStorage) GetAllCountersMap(_ context.Context) (map[string]int64, error) {
	return ms.CounterMap, nil
}

func (ms MemStorage) GetAllMetrics(_ context.Context) (any, error) {
	return ms, nil
}

func (ms MemStorage) Close() error {
	return nil
}
