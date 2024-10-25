package memstorage

import (
	"errors"
)

// MemStorage inmemory хранилище для метрик. Разные map-ы для разных типов метрик
type MemStorage struct {
	GaugeMap    map[string]float64
	CounterName map[string]int64
}

// func New() storage.Storager {
func New() MemStorage {
	return MemStorage{
		GaugeMap:    make(map[string]float64),
		CounterName: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateGauge(key string, value float64) error {
	ms.GaugeMap[key] = value
	return nil
}

func (ms *MemStorage) UpdateCounter(key string, value int64) error {
	ms.CounterName[key] += value
	return nil
}

func (ms *MemStorage) GetGauge(key string) (float64, error) {
	val, ok := ms.GaugeMap[key]
	//log.Println("GetGauge key:", key, "val:", val, "ok:", ok)
	if !ok {
		return 0, errors.New("no value for key " + key)
	}
	return val, nil
}

func (ms *MemStorage) GetCounter(key string) (int64, error) {
	val, ok := ms.CounterName[key]
	if !ok {
		return 0, errors.New("no value for key " + key)
	}
	return val, nil
}

func (ms *MemStorage) GetValue(t string, key string) (any, error) {
	if t == "counter" {
		val, ok := ms.CounterName[key]
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

func (ms *MemStorage) GetAllGaugesMap() (map[string]float64, error) {
	return ms.GaugeMap, nil
}

func (ms *MemStorage) GetAllCountersMap() (map[string]int64, error) {
	return ms.CounterName, nil
}

func (ms *MemStorage) GetAllMetrics() (any, error) {
	return *ms, nil
}
