package storage

type Metrics struct {
	ID    string   `json:"id"`              // Имя метрики
	MType string   `json:"type"`            // Параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // Значение метрики в случае передачи gauge
}

//// Storager интерфейс, устанавливает контракт для хранилища метрик вне зависимости от конкретной реализации хранилища
//type Storager interface {
//	UpdateGauge(key string, value float64) error
//	UpdateCounter(key string, value int64) error
//	GetGauge(key string) (float64, error)
//	GetCounter(key string) (int64, error)
//	GetValue(t string, key string) (any, error)
//	GetAllGaugesMap() (map[string]float64, error)
//	GetAllCountersMap() (map[string]int64, error)
//	GetAllMetrics() (any, error)
//}
