package internal

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"logger/cmd/server/initconf"
	"logger/internal/handlers"
	"logger/internal/storage/memstorage"
	"os"
)

type Storager interface {
	GetAllMetrics(ctx context.Context) (any, error)
}

// Save функция сохранения дампа метрик в файл.
func Save(ctx context.Context, store handlers.Storager, fname string) error {
	// сериализуем структуру в JSON формат
	metrics, err := store.GetAllMetrics(ctx)
	if err != nil {
		log.Println("error store serialisation in Save", err)
		return err
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		log.Println("Save. Error marshalling store")
		return err
	}

	err = os.WriteFile(fname, data, 0666)
	if err != nil {
		log.Println("Save. Error os.WriteFile")
		return err
	}
	return nil
}

// Load функция чтения дампа метрик из файла. Применимо только для memstorage
// func Load(_ context.Context, store handlers.Storager, fname string) error {
func Load(store *handlers.Storager, fname string) error {

	// Временное хранилище для Unmarshall-инга в необходимую структуру memstorage
	var memStore memstorage.MemStorage
	data, err := os.ReadFile(fname)
	if err != nil {
		print("Save. Error read store dump file", fname)
		return err
	}
	//err = json.Unmarshal(data, &store)
	err = json.Unmarshal(data, &memStore)
	if err != nil {
		log.Println("Load. Error unmarshalling from file")
		return err
	}
	*store = memStore
	log.Println("storage from Load:", store)
	return nil
}

// SyncDumpUpdate middleware для апдейта файла дампа метрик каждый раз при приходе новой метрики
// Для случая ключа STORE_INTERVAL = 0
func SyncDumpUpdate(ctx context.Context, store handlers.Storager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		log.Println("SyncDumpUpdate StoreMetricInterval :", initconf.Conf.StoreMetricInterval)
		if initconf.Conf.StoreMetricInterval == 0 {
			log.Println("sync flush metric into dump")
			if err := Save(ctx, store, initconf.Conf.FileStoragePath); err != nil {
				log.Println("SyncDumpUpdate error:", err)
			}
		}
	}
}
