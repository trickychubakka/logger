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
	mygzip "logger/internal/compress"
	"logger/internal/handlers"
	"logger/internal/logging"
	"os"
	"time"
)

// Для возможности использования Zap
var sugar zap.SugaredLogger

// task функция для старта дампа метрик на диск
// func task(ctx context.Context, interval int, store *memstorage.MemStorage) {
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

var ctx, ctxPG, ctxDUMP context.Context
var cancel, cancelPG, cancelDUMP context.CancelFunc
var store handlers.Storager

func main() {

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// Config initialization
	if err := initconf.InitConfig(&initconf.Conf); err != nil {
		log.Println("Panic in initConfig")
		panic(err)
	}
	log.Println("initconf is:", initconf.Conf)

	var err error
	if initconf.Conf.DatabaseDSN == "" {
		// Store Инициализация хранилища метрик
		store, err = memstorage.New(ctx)
		if err != nil {
			log.Println("Error memstorage initialization.")
			panic(err)
		}
	}
	if initconf.Conf.DatabaseDSN != "" {
		store, err = pgstorage.New(ctx)
		if err != nil {
			log.Println("Error pgstorage initialization.")
			panic(err)
		}
	}

	defer store.Close()

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
	// делаем регистратор SugaredLogger
	sugar = *logger.Sugar()

	if initconf.Conf.DatabaseDSN == "" && initconf.Conf.Restore {
		if err := internal.Load(&store, initconf.Conf.FileStoragePath); err != nil {
			log.Println("Error in initial dump load:", err)
		}
	}

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

	//if initconf.Conf.StoreMetricInterval != 0 {
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

	// Изменение режима работы GIN
	//gin.SetMode(gin.ReleaseMode)

	if initconf.Conf.Logfile != "" {
		file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file:", err)
		}
		log.SetOutput(file)
		defer file.Close()
	}

	log.Println("TEST Connecting to base")
	db := database.Postgresql{}
	err = db.Connect()
	if err != nil {
		log.Println("Error connecting to database :", err)
	}
	defer db.Close()

	router := gin.Default()

	router.Use(logging.WithLogging(&sugar))
	router.Use(gzip.Gzip(gzip.DefaultCompression)) //-- standard GIN compress "github.com/gin-contrib/compress"

	router.Use(mygzip.GzipRequestHandle)
	//router.Use(mygzip.GzipResponseHandle(compress.DefaultCompression))
	if initconf.Conf.DatabaseDSN == "" {
		router.Use(internal.SyncDumpUpdate(ctx, store))
	}
	router.GET("/", handlers.GetAllMetrics(ctx, store))
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler(ctx, store))
	router.POST("/update", handlers.MetricHandlerJSON(ctx, store))
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
