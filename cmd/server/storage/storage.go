package storage

type Storager interface {
	UpdateGauge(key string, value float64) error
	UpdateCounter(key string, value int64) error
}
