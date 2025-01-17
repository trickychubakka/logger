// Iter5 branch
package main

import (
	"context"
	"logger/internal"
	"logger/internal/storage/memstorage"
	"logger/internal/storage/pgstorage"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
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
func task(ctx context.Context, interval int, store handlers.Storager, conf *initconf.Config) {
	// запускаем бесконечный цикл
	for {
		select {
		// проверяем не завершён ли ещё контекст и выходим, если завершён
		case <-ctx.Done():
			return

		// выполняем нужный нам код
		default:
			println("Save metrics dump to file", conf.FileStoragePath, "with interval", interval, "s")
			err := internal.Save(ctx, store, conf.FileStoragePath)
			if err != nil {
				return
			}
		}
		// делаем паузу перед следующей итерацией
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// storeInit функция инициализации store. В зависимости от настроек (env, флаги) будет либо
// 1. создан memstorage восстановлением из dump-а
// 2. при ошибке в п.1 -- создан новый memstorage
// 3. если определена переменная DatabaseDSN -- создан store типа pgstorage
func storeInit(ctx context.Context, store handlers.Storager, conf *initconf.Config) (handlers.Storager, error) {
	var err error
	if conf.DatabaseDSN == "" {
		// если определена опция восстановления store из дампа
		if conf.Restore {
			log.Println("DatabaseDSN is not configured, Load metric dump from file")
			store, err := internal.Load(conf.FileStoragePath)
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
			return nil, err
		}
		return store, nil
	}
	// Инициализация хранилища метрик типа pgstorage
	if conf.DatabaseDSN != "" {
		log.Println("storeInit DatabaseDSN is configured, start to initialize pgstorage.")
		store, err = pgstorage.New(ctx, conf)
		if err != nil {
			log.Println("storeInit error pgstorage initialization.")
			return nil, err
		}
	}
	return store, nil
}

var err error

func main() {
	// Изменение режима работы GIN
	//gin.SetMode(gin.ReleaseMode)

	var ctx, ctxDUMP context.Context
	var cancel, cancelDUMP context.CancelFunc
	var conf initconf.Config
	var store handlers.Storager

	// Parent context
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// Config initialization
	if err := initconf.InitConfig(&conf); err != nil {
		log.Println("Panic in initConfig")
		panic(err)
	}
	log.Println("initconf is:", conf)

	// store initialization
	if store, err = storeInit(ctx, store, &conf); err != nil {
		log.Println("Storage initialization error :", err)
		panic(err)
	}
	defer store.Close()

	// Остановка сервера и сохранение дампа memstorage при остановке, если используется memstorage
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// Если для хранения метрик не используется БД -- делаем DUMP метрик на диск
		if conf.DatabaseDSN == "" {
			err := internal.Save(ctx, store, conf.FileStoragePath)
			if err != nil {
				log.Println("Save metric DUMP error:", err)
			}
		}
		log.Println("SERVER STOPPED.")
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

	// Если не определен DatabaseDSN и StoreMetricInterval не равен нулю -- запускается автодамп memstorage
	if conf.DatabaseDSN == "" && conf.StoreMetricInterval != 0 {
		// создаём контекст с функцией завершения
		log.Println("Init context fo goroutine (Conf.StoreMetricInterval is not 0):", conf.StoreMetricInterval)
		// Создаем дочерний контекст для процесса дампа метрик в случае, если StoreMetricInterval != 0
		ctxDUMP, cancelDUMP = context.WithCancel(ctx)
		// запускаем горутину
		go task(ctxDUMP, conf.StoreMetricInterval, store, &conf)
	}

	sugar.Infow("initConfig sugar logging", "conf.RunAddr", conf.RunAddr)

	// Если определена опция Logfile -- логи сервера перенаправляются в этот файл
	if conf.Logfile != "" {
		file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file:", err)
		}
		log.SetOutput(file)
		defer file.Close()
	}

	// GIN init
	router := gin.Default()

	router.Use(logging.WithLogging(&sugar))
	router.Use(gzip.Gzip(gzip.DefaultCompression)) //-- standard GIN compress "github.com/gin-contrib/compress"

	router.Use(compress.GzipRequestHandle(ctx, &conf))
	//router.Use(compress.GzipResponseHandle(compress.DefaultCompression))
	if conf.DatabaseDSN == "" {
		router.Use(internal.SyncDumpUpdate(ctx, store, &conf))
	}
	router.GET("/", handlers.GetAllMetrics(ctx, store))
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler(ctx, store))
	router.POST("/update/", handlers.MetricHandlerJSON(ctx, store, &conf))
	router.POST("/updates", handlers.MetricHandlerBatchUpdate(ctx, store, &conf))
	router.GET("/value/:metricType/:metricName", handlers.GetMetric(ctx, store))
	router.POST("/value/", handlers.GetMetricJSON(ctx, store, &conf))
	router.GET("/ping", handlers.DBPing(conf.DatabaseDSN))

	// Start PProf HTTP if option -t enabled
	if conf.PProfHTTPEnabled {
		pprof.Register(router)
	}

	err = router.Run(conf.RunAddr)
	if err != nil {
		panic(err)
	}

	sugar.Infow("\nServer started on runAddr %s \n", conf.RunAddr)

	// завершаем дочерний контекст дампа, чтобы завершить горутину дампа метрик в файл
	if conf.DatabaseDSN == "" && conf.StoreMetricInterval != 0 {
		cancelDUMP()
	}

	log.Println("SERVER STOPPED.")
}
