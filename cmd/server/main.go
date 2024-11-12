// Iter5 branch
package main

import (
	"context"
	"logger/internal"
	"logger/internal/database"
	"logger/internal/storage/memstorage"
	"logger/internal/storage/pgstorage"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"logger/cmd/server/initconf"
	"logger/internal/compress"
	"logger/internal/handlers"
	"logger/internal/logging"
	"os"
	"time"
)

// Для возможности использования Zap
var sugar zap.SugaredLogger

// task функция для старта дампа метрик на диск раз в interval секунд
func task(ctx context.Context, interval int, store handlers.Storager) {
	// запускаем бесконечный цикл
	for {
		select {
		// проверяем не завершён ли ещё контекст и выходим, если завершён
		case <-ctx.Done():
			return

		// выполняем нужный нам код
		default:
			println("Save metrics dump to file", initconf.Conf.FileStoragePath, "with interval", interval, "s")
			err := internal.Save(ctx, store, initconf.Conf.FileStoragePath)
			if err != nil {
				return
			}
		}
		// делаем паузу перед следующей итерацией
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// Контексты: родительский, контекст postgres, контекст DUMP-а
var ctx, ctxPG, ctxDUMP context.Context
var cancel, cancelPG, cancelDUMP context.CancelFunc

// storeInit функция инициализации store. В зависимости от настроек (env, флаги) будет либо
// 1. создан memstorage восстановлением из dump-а
// 2. при ошибке в п.1 -- создан новый memstorage
// 3. если определена переменная DatabaseDSN -- создан store типа pgstorage
func storeInit(ctx context.Context) (handlers.Storager, error) {
	var err error
	//var store handlers.Storager
	if initconf.Conf.DatabaseDSN == "" {
		// если определена опция восстановления store из дампа
		if initconf.Conf.Restore {
			log.Println("DatabaseDSN is not configured, Load metric dump from file")
			store, err := internal.Load(initconf.Conf.FileStoragePath)
			log.Println("Metric after dump load :", store)
			if err == nil {
				return store, nil
			} else {
				log.Println("storeInit error in initial dump load:", err, " Trying to initialize new memstorage.")
			}
		}
		// Store Инициализация хранилища метрик типа memstorage
		store, err = memstorage.New(ctx)
		if err != nil {
			log.Println("storeInit error memstorage initialization.")
			panic(err)
		}
		return store, nil
	}
	// Инициализация хранилища метрик типа pgstorage
	if initconf.Conf.DatabaseDSN != "" {
		log.Println("storeInit DatabaseDSN is configured, start to initialize pgstorage.")
		store, err = pgstorage.New(ctx)
		if err != nil {
			log.Println("storeInit error pgstorage initialization.")
			panic(err)
		}
	}
	return store, nil
}

var store handlers.Storager
var err error

func main() {
	// Изменение режима работы GIN
	//gin.SetMode(gin.ReleaseMode)

	// Parent context
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// Config initialization
	if err := initconf.InitConfig(&initconf.Conf); err != nil {
		log.Println("Panic in initConfig")
		panic(err)
	}
	log.Println("initconf is:", initconf.Conf)

	// store initialization
	if store, err = storeInit(ctx); err != nil {
		log.Println("Storage initialization error :", err)
		panic(err)
	}
	defer store.Close()

	// Сохранение дампа memstorage при остановке
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := internal.Save(ctx, store, initconf.Conf.FileStoragePath)
		if err != nil {
			return
		}
		log.Println("SERVER STOPPED!!!")
		os.Exit(1)
	}()

	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	// создаем регистратор SugaredLogger
	sugar = *logger.Sugar()

	/*
		// PostgreSQL Store инициализация
		ctxPG, cancelPG = context.WithCancel(ctx)
		defer cancelPG()
		pgStore, err := pgstorage.New(ctxPG)
		if err != nil {
			log.Println("pgstorage package error in New():", err)
		}
		log.Println("pgStore initialization:", pgStore)
		//ctx := context.Background()
		pgStore.UpdateGauge(ctxPG, "value1", 2.3)
		pgStore.UpdateCounter(ctxPG, "value2", 3)
		if tmp, err := pgStore.GetGauge(ctxPG, "value1"); err == nil {
			log.Println("PG GetGauge(value1) :", tmp)
		} else {
			log.Println("PG GetGauge(value1) error:", err)
		}
		if tmp, err := pgStore.GetCounter(ctxPG, "value2"); err == nil {
			log.Println("PG GetCounter(value2) :", tmp)
		}
		if tmp, err := pgStore.GetValue(ctxPG, "counter", "value2"); err == nil {
			log.Println("pgStore.GetValue(counter, value2) :", tmp)
		}
		if tmp, err := pgStore.GetValue(ctxPG, "gauge", "value1"); err == nil {
			log.Println("pgStore.GetValue(gauge, value1) :", tmp)
		}
		if tmp, err := pgStore.GetValue(ctx, "gauge", "errmetric"); err != nil {
			log.Println("pgStore.GetValue(gauge, errmetric) :", tmp, "Error:", err)
		}
		if m, err := pgStore.GetAllMetrics(ctxPG); err == nil {
			log.Println("pgStore.GetAllMetrics():", m)
		} else {
			log.Println("pgStore.GetAllMetrics() error:", err)
		}
	*/

	// Если не определен DatabaseDSN и StoreMetricInterval не равен нулю -- запускается автодамп memstorage
	if initconf.Conf.DatabaseDSN == "" && initconf.Conf.StoreMetricInterval != 0 {
		// создаём контекст с функцией завершения
		log.Println("Init context fo goroutine (Conf.StoreMetricInterval is not 0):", initconf.Conf.StoreMetricInterval)
		//ctx, cancel = context.WithCancel(context.Background())
		// Создаем дочерний контекст для процесса дампа метрик в случае, если StoreMetricInterval != 0
		ctxDUMP, cancelDUMP = context.WithCancel(ctx)
		// запускаем горутину
		go task(ctxDUMP, initconf.Conf.StoreMetricInterval, store)
	}

	sugar.Infow("initConfig sugar logging", "conf.RunAddr", initconf.Conf.RunAddr)

	// Если определена опция Logfile -- логи сервера перенаправляются в этот файл
	if initconf.Conf.Logfile != "" {
		file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file:", err)
		}
		log.SetOutput(file)
		defer file.Close()
	}

	// Тест коннекта к базе
	log.Println("TEST Connecting to base")
	db := database.Postgresql{}
	err = db.Connect()
	if err != nil {
		log.Println("Error connecting to database :", err)
	}
	defer db.Close()

	// GIN init
	router := gin.Default()

	router.Use(logging.WithLogging(&sugar))
	router.Use(gzip.Gzip(gzip.DefaultCompression)) //-- standard GIN compress "github.com/gin-contrib/compress"

	router.Use(compress.GzipRequestHandle)
	//router.Use(compress.GzipResponseHandle(compress.DefaultCompression))
	if initconf.Conf.DatabaseDSN == "" {
		router.Use(internal.SyncDumpUpdate(ctx, store))
	}
	router.GET("/", handlers.GetAllMetrics(ctx, store))
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler(ctx, store))
	router.POST("/update", handlers.MetricHandlerJSON(ctx, store))
	router.POST("/updates", handlers.MetricHandlerBatchUpdate(ctx, store))
	router.GET("/value/:metricType/:metricName", handlers.GetMetric(ctx, store))
	router.POST("/value", handlers.GetMetricJSON(ctx, store))
	router.GET("/ping", handlers.DBPing(db.DB))

	err = router.Run(initconf.Conf.RunAddr)
	if err != nil {
		panic(err)
	}

	sugar.Infow("\nServer started on runAddr %s \n", initconf.Conf.RunAddr)

	// завершаем дочерний контекст дампа, чтобы завершить горутину дампа метрик в файл
	if initconf.Conf.StoreMetricInterval != 0 {
		cancelDUMP()
	}

	log.Println("SERVER STOP!!!")
}
