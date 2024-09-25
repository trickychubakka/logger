package memstorage

import (
	"storage"
)

// MemStorage inmemory хранилище для метрик. Разные map-ы для разных типов метрик
type MemStorage struct {
	gaugeMap   map[string]float64
	counterMap map[string]int64
}

func New() storage.Storager {
	return &MemStorage{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateGauge(key string, value float64) error {
	ms.gaugeMap[key] = value
	return nil
}

func (ms *MemStorage) UpdateCounter(key string, value int64) error {
	ms.counterMap[key] += value
	return nil
}
