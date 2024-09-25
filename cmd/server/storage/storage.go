package storage

// Storager интерфейс, устанавливает контракт для хранилища метрик вне зависимости от конкретной реализации хранилища
type Storager interface {
	UpdateGauge(key string, value float64) error
	UpdateCounter(key string, value int64) error
}
