package internal

import (
	"encoding/json"
	"log"
	"logger/cmd/server/initconfig"
	"logger/internal/storage/memstorage"
	"os"
)

//var storageFile = memstorage.New()

// Save сохраняет настройки в файле fname.
func Save(store *memstorage.MemStorage, fname string) error {
	// сериализуем структуру в JSON формат
	metrics, err := store.GetAllMetrics()
	if err != nil {
		log.Println("error store serialisation in Save", err)
		return err
	}
	//log.Println("Store in Save:", metrics)
	data, err := json.Marshal(metrics)
	if err != nil {
		log.Println("Save. Error marshalling store")
		//return err
	}
	//log.Println("data", data)
	// сохраняем данные в файл
	os.WriteFile(initconfig.Conf.FileStoragePath, data, 0666)
	return nil
}

// Load читает настройки из файла fname.
func Load(store *memstorage.MemStorage, fname string) error {
	// прочитайте файл с помощью os.ReadFile
	// десериализуйте данные используя json.Unmarshal
	// ...
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
