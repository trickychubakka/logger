// Package storage -- пакет с определением объектов используемого в logger хранилища.
package storage

// Metrics структура для хранения одной метрики.
type Metrics struct {
	ID    string   `json:"id"`              // Имя метрики.
	MType string   `json:"type"`            // Параметр, принимающий значение gauge или counter.
	Delta *int64   `json:"delta,omitempty"` // Значение метрики в случае передачи counter.
	Value *float64 `json:"value,omitempty"` // Значение метрики в случае передачи gauge.
}
