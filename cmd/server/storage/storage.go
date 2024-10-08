package storage

// Storager интерфейс, устанавливает контракт для хранилища метрик вне зависимости от конкретной реализации хранилища
type Storager interface {
	UpdateGauge(key string, value float64) error
	UpdateCounter(key string, value int64) error
	GetGauge(key string) (float64, error)
	GetCounter(key string) (int64, error)
	GetValue(t string, key string) (any, error)
	GetAllGaugesMap() (map[string]float64, error)
	GetAllCountersMap() (map[string]int64, error)
}
