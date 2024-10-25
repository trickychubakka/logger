// Iter5 branch
package main

import (
	"context"
	"logger/internal"
	"logger/internal/storage/memstorage"
	"os/signal"
	"syscall"

	//"compress/gzip"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"logger/cmd/server/initconf"
	mygzip "logger/internal/gzip"
	"logger/internal/handlers"
	"logger/internal/logging"
	"os"
	"time"
)

// var conf Config

// Для возможности использования Zap -- в каждом модуле определять?
var sugar zap.SugaredLogger

//
//// FlagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
//var FlagTest = false

//type tcpKeepAliveListener struct {
//	*net.TCPListener
//}
//
//func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
//	tc, err := ln.AcceptTCP()
//	if err != nil {
//		return
//	}
//	tc.SetKeepAlive(true)
//	tc.SetKeepAlivePeriod(3 * time.Minute)
//	return tc, nil
//}

// task функция для старта дампа метрик на диск
func task(ctx context.Context, interval int, store *memstorage.MemStorage) {
	// запускаем бесконечный цикл
	for {
		select {
		// проверяем не завершён ли ещё контекст и выходим, если завершён
		case <-ctx.Done():
			return

		// выполняем нужный нам код
		default:
			println("Save metrics dump to file", initconf.Conf.FileStoragePath, "with interval", interval, "s")
			err := internal.Save(store, initconf.Conf.FileStoragePath)
			if err != nil {
				return
			}
		}
		// делаем паузу перед следующей итерацией
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

var ctx context.Context
var cancel context.CancelFunc

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		//cleanup()
		//os.WriteFile("interrupt.dump", []byte("TEST!"), 0666)
		err := internal.Save(&initconf.Store, initconf.Conf.FileStoragePath)
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

	if err := initconf.InitConfig(&initconf.Conf); err != nil {
		log.Println("Panic in initConfig")
		panic(err)
	}
	log.Println("initconf is:", initconf.Conf)

	//defer os.WriteFile("interrupt.dump", []byte("TEST!"), 0666)
	defer internal.Save(&initconf.Store, initconf.Conf.FileStoragePath)

	if initconf.Conf.Restore {
		if err := internal.Load(&initconf.Store, initconf.Conf.FileStoragePath); err != nil {
			log.Println("Error in initial dump load:", err)
		}
	}

	if initconf.Conf.StoreMetricInterval != 0 {
		// создаём контекст с функцией завершения
		log.Println("Init context fo goroutine (Conf.StoreMetricInterval is not 0):", initconf.Conf.StoreMetricInterval)
		ctx, cancel = context.WithCancel(context.Background())
		// запускаем горутину
		go task(ctx, initconf.Conf.StoreMetricInterval, &initconf.Store)
	}

	sugar.Infow("initConfig sugar logging", "conf", initconf.Conf.RunAddr)
	//gin.SetMode(gin.ReleaseMode)

	if initconf.Conf.Logfile != "" {
		file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file:", err)
		}
		log.SetOutput(file)
		defer file.Close()
	}

	router := gin.Default()
	//router.Use(limit.Limit(200))
	router.Use(logging.WithLogging(&sugar))
	router.Use(gzip.Gzip(gzip.DefaultCompression)) //-- standard GIN gzip "github.com/gin-contrib/gzip"
	//router.Use(gin.Recovery())
	//custom gzip handlers
	router.Use(mygzip.GzipRequestHandle)
	//router.Use(mygzip.GzipResponseHandle(gzip.DefaultCompression))
	router.Use(internal.SyncDumpUpdate())
	router.GET("/", handlers.GetAllMetrics)
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler)
	router.POST("/update", handlers.MetricHandlerJSON)
	router.GET("/value/:metricType/:metricName", handlers.GetMetric)
	router.POST("/value", handlers.GetMetricJSON)

	err = router.Run(initconf.Conf.RunAddr)
	if err != nil {
		panic(err)
	}

	sugar.Infow("\nServer started on runAddr %s \n", initconf.Conf.RunAddr)

	// завершаем контекст, чтобы завершить горутину дампа метрик в файл
	if initconf.Conf.StoreMetricInterval != 0 {
		cancel()
	}

	log.Println("SERVER STOP!!!")
}
