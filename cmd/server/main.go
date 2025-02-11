// Logger сервер приема и хранения метрик.

package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"logger/cmd/server/initconf"
	"logger/config"
	"logger/internal"
	"logger/internal/compress"
	"logger/internal/encryption"
	"logger/internal/handlers"
	"logger/internal/logging"
	"logger/internal/storage/memstorage"
	"logger/internal/storage/pgstorage"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	//configFile = `C:\JetBrains\GolandProjects\logger\internal\config\server.json`
	configFile      = `./config/server.json`
	shutdownTimeout = 5
)

// Для возможности использования Zap.
var sugar zap.SugaredLogger

var buildVersion, buildDate, buildCommit string

// task функция дампа метрик на диск раз в interval секунд.
func task(ctx context.Context, interval int, store handlers.Storager, conf *config.Config) {
	// запускаем бесконечный цикл
	for {
		select {
		// проверяем не завершён ли ещё контекст и выходим, если завершён.
		case <-ctx.Done():
			return
		// выполняем нужный нам код.
		default:
			println("Save metrics dump to file", conf.FileStoragePath, "with interval", interval, "s")
			err := internal.Save(ctx, store, conf.FileStoragePath)
			if err != nil {
				return
			}
		}
		// делаем паузу перед следующей итерацией.
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// storeInit функция инициализации store. В зависимости от настроек (env, флаги) будет принят один из следующих вариантов.
// 1. создан memstorage восстановлением из dump-а
// 2. при ошибке в п.1 -- создан новый memstorage
// 3. если определена переменная DatabaseDSN -- создан store типа pgstorage
func storeInit(ctx context.Context, store handlers.Storager, conf *config.Config) (handlers.Storager, error) {
	var err error
	if conf.DatabaseDSN == "" {
		// если определена опция восстановления store из дампа.
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
		// Инициализация хранилища метрик типа memstorage.
		store, err = memstorage.New(ctx)
		if err != nil {
			log.Println("storeInit error memstorage initialization.")
			return nil, err
		}
		return store, nil
	}
	// Инициализация хранилища метрик типа pgstorage.
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
	// Изменение режима работы GIN.
	//gin.SetMode(gin.ReleaseMode)

	internal.PrintStartMessage(buildVersion, buildDate, buildCommit)

	var ctx, ctxDUMP context.Context
	var cancel, cancelDUMP context.CancelFunc
	var conf config.Config
	var store handlers.Storager

	// Parent context.
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	err = config.ReadConfig(configFile, &conf)
	if err != nil {
		log.Println("Config pre-initialization from config file", configFile, " error :", err)
	}
	log.Println("Pre-initialized from server.json config is :", fmt.Sprintf("%+v\n", conf))

	// Config initialization.
	if err = initconf.InitConfig(&conf); err != nil {
		log.Println("Panic in initConfig")
		panic(err)
	}
	log.Println("initconf is:", fmt.Sprintf("%+v\n", conf))

	// Store initialization.
	if store, err = storeInit(ctx, store, &conf); err != nil {
		log.Println("Storage initialization error :", err)
		panic(err)
	}
	defer store.Close()

	// Создаём предустановленный регистратор zap.
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	// Создаем регистратор SugaredLogger
	sugar = *logger.Sugar()

	// Если не определен DatabaseDSN и StoreMetricInterval не равен нулю -- запускается автодамп memstorage.
	if conf.DatabaseDSN == "" && conf.StoreMetricInterval != 0 {
		// Создаём контекст с функцией завершения.
		log.Println("Init context fo goroutine (Conf.StoreMetricInterval is not 0):", conf.StoreMetricInterval)
		// Создаем дочерний контекст для процесса дампа метрик в случае, если StoreMetricInterval != 0
		ctxDUMP, cancelDUMP = context.WithCancel(ctx)
		// Запускаем горутину.
		go task(ctxDUMP, conf.StoreMetricInterval, store, &conf)
	}

	sugar.Infow("initConfig sugar logging", "conf.RunAddr", conf.RunAddr)

	// Если определена опция Logfile -- логи сервера перенаправляются в этот файл.
	if conf.Logfile != "" {
		file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file:", err)
		}
		log.SetOutput(file)
		defer file.Close()
	}

	// GIN init.
	router := gin.Default()
	router.Use(logging.WithLogging(&sugar))
	if conf.TrustedSubnet != "" {
		router.Use(handlers.CheckTrustedSubnet(&conf))
	}
	if conf.PathToPrivateKey != "" {
		router.Use(encryption.DecryptRequestHandler(ctx, conf.PrivateKey))
	}
	router.Use(gzip.Gzip(gzip.DefaultCompression)) //-- standard GIN compress "github.com/gin-contrib/compress".
	router.Use(compress.GzipRequestHandle(ctx, &conf))
	if conf.DatabaseDSN == "" {
		router.Use(internal.SyncDumpUpdate(ctx, store, &conf))
	}
	// для обработки всех запросов, не обрабатываемых handler-ами ниже -- из-за gzip handler проблемы.
	router.NoRoute(func(c *gin.Context) {
		fmt.Println("Handling Any request for", c.Request.URL.Path)
		c.Status(http.StatusNotFound)
		// Flush -- убираем возможность изменения статуса gzip handler-ом:
		c.Writer.Flush()
	})
	router.GET("/", handlers.GetAllMetrics(ctx, store))
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler(ctx, store))
	router.POST("/update/", handlers.MetricHandlerJSON(ctx, store, &conf))
	router.POST("/updates", handlers.MetricHandlerBatchUpdate(ctx, store, &conf))
	router.GET("/value/:metricType/:metricName", handlers.GetMetric(ctx, store))
	router.POST("/value/", handlers.GetMetricJSON(ctx, store, &conf))
	router.GET("/ping", handlers.DBPing(conf.DatabaseDSN))

	// Start PProf HTTP if option -t enabled.
	if conf.PProfHTTPEnabled {
		pprof.Register(router)
	}

	// Запуск сервера через http.Server, чтобы воспользоваться Shutdown методом для graceful shutdown.
	srv := &http.Server{
		Addr:    conf.RunAddr,
		Handler: router.Handler(),
	}

	// Запуск listener в горутине, чтобы не блокировать graceful shutdown ниже.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	sugar.Infow("\nServer started on runAddr ", "RunAddr", conf.RunAddr)

	// Остановка сервера и сохранение дампа memstorage при остановке, если используется memstorage.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	log.Println("Shutdown Server ...")
	// Если для хранения метрик не используется БД -- делаем DUMP метрик на диск.
	if conf.DatabaseDSN == "" {
		// Завершаем дочерний контекст дампа, чтобы завершить горутину дампа метрик в файл.
		if conf.StoreMetricInterval != 0 {
			cancelDUMP()
		}
		err := internal.Save(ctx, store, conf.FileStoragePath)
		if err != nil {
			log.Println("Save metric DUMP error:", err)
		}
	}

	ctxHTTP, cancelHTTP := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
	defer cancelHTTP()
	if err := srv.Shutdown(ctxHTTP); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	<-ctxHTTP.Done()
	log.Println("timeout of", shutdownTimeout, " seconds.")

	log.Println("SERVER STOPPED.")
}
