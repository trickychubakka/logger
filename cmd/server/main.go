// Iter5 branch
package main

import (
	//"compress/gzip"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	mygzip "logger/internal/gzip"
	"logger/internal/handlers"
	"logger/internal/logging"
	"os"
)

var conf Config
var sugar zap.SugaredLogger

// FlagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
var FlagTest = false

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

func main() {

	//var conf Config

	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	// делаем регистратор SugaredLogger
	sugar = *logger.Sugar()

	if err := initConfig(&conf); err != nil {
		log.Println("Panic in initConfig")
		panic(err)
	}
	sugar.Infow("initConfig sugar logging", "conf", conf.runAddr)
	//gin.SetMode(gin.ReleaseMode)

	if conf.logfile != "" {
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
	//custom gzip handler
	router.Use(mygzip.GzipRequestHandle)
	//router.Use(mygzip.GzipResponseHandle(gzip.DefaultCompression))
	router.GET("/", handlers.GetAllMetrics)
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler)
	router.POST("/update", handlers.MetricHandlerJSON)
	router.GET("/value/:metricType/:metricName", handlers.GetMetric)
	router.POST("/value", handlers.GetMetricJSON)

	err = router.Run(conf.runAddr)
	if err != nil {
		panic(err)
	}

	// ? start
	//server := &http.Server{Handler: router}
	//l, err := net.Listen("tcp4", conf.runAddr)
	//if err != nil {
	//	log.Fatal(err)
	//}
	////err = server.Serve(l)
	//
	//server.Serve(tcpKeepAliveListener{l.(*net.TCPListener)})
	// ? end

	sugar.Infow("\nServer started on runAddr %s \n", conf.runAddr)
}
