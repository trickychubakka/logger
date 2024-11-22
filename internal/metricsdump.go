package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"logger/cmd/server/initconf"
	"logger/internal/handlers"
	"logger/internal/storage/memstorage"
	"os"
	"time"
)

type Storager interface {
	GetAllMetrics(ctx context.Context) (any, error)
}

// Save функция сохранения дампа метрик в файл.
func Save(ctx context.Context, store Storager, fname string) error {
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
		// константа timeoutsRetryConst = [3]int{1, 3, 5} определена в metrics.go
		for i, t := range timeoutsRetryConst {
			log.Println("Save: Trying to recover after ", t, "seconds, attempt number ", i+1)
			time.Sleep(time.Duration(t) * time.Second)
			err := os.WriteFile(fname, data, 0666)
			if err != nil {
				log.Println("Save: attempt ", i+1, " error")
				if i == 2 {
					return fmt.Errorf("%s %v", "Save: os.WriteFile error:", err)
				}
				continue
			}
		}
	}
	return nil
}

// Load функция чтения дампа метрик из файла. Применимо только для memstorage
func Load(fname string) (handlers.Storager, error) {
	var store handlers.Storager
	// Временное хранилище для Unmarshall-инга в необходимую структуру memstorage
	var memStore memstorage.MemStorage
	data, err := os.ReadFile(fname)
	if err != nil {
		log.Println("Save. Error read store dump file", fname)
		return nil, err
	}
	// Использование метода Unmarshal пакета memstorage из-за непубличности полей, аналог вызова err = json.Unmarshal(data, &memStore)
	err = memstorage.Unmarshal(data, &memStore)
	if err != nil {
		log.Println("Load. Error unmarshalling from file")
		return nil, err
	}
	store = memStore
	log.Println("storage from Load:", store)
	return store, nil
}

// SyncDumpUpdate middleware для апдейта файла дампа метрик каждый раз при приходе новой метрики
// Для случая ключа STORE_INTERVAL = 0
func SyncDumpUpdate(ctx context.Context, store handlers.Storager, conf *initconf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		log.Println("SyncDumpUpdate StoreMetricInterval :", conf.StoreMetricInterval)
		if conf.StoreMetricInterval == 0 {
			log.Println("sync flush metric into dump")
			if err := Save(ctx, store, conf.FileStoragePath); err != nil {
				log.Println("SyncDumpUpdate error:", err)
			}
		}
	}
}
