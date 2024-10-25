package internal

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"logger/cmd/server/initconf"
	"logger/internal/storage/memstorage"
	"os"
)

//var storageFile = memstorage.New()

// Save функция сохранения дампа метрик в файл.
func Save(store *memstorage.MemStorage, fname string) error {
	// сериализуем структуру в JSON формат
	metrics, err := store.GetAllMetrics()
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

// Load функция чтения дампа метрик из файла
func Load(store *memstorage.MemStorage, fname string) error {

	data, err := os.ReadFile(fname)
	if err != nil {
		print("Save. Error read store dump file", fname)
		return err
	}
	err = json.Unmarshal(data, &store)
	if err != nil {
		log.Println("Load. Error unmarshalling from file")
		return err
	}
	log.Println("storage from Load:", store)
	return nil
}

// SyncDumpUpdate middleware для апдейта файла дампа метрик каждый раз при приходе новой метрики
// Для случая ключа STORE_INTERVAL = 0
func SyncDumpUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		log.Println("SyncDumpUpdate StoreMetricInterval :", initconf.Conf.StoreMetricInterval)
		if initconf.Conf.StoreMetricInterval == 0 {
			log.Println("sync flush metric into dump")
			if err := Save(&initconf.Store, initconf.Conf.FileStoragePath); err != nil {
				log.Println("SyncDumpUpdate error:", err)
			}
		}
		//c.Next()
	}
}
