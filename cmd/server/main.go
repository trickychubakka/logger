// Iter5 branch
package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"logger/internal/handlers"
	"logger/internal/logging"
)

var conf Config
var sugar zap.SugaredLogger

// FlagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
var FlagTest = false

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
	sugar.Infow("initConfig sugar logging", "conf", conf)
	//gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(logging.WithLogging(&sugar))
	router.GET("/", handlers.GetAllMetrics)
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler)
	router.POST("/update", handlers.MetricHandlerJSON)
	router.GET("/value/:metricType/:metricName", handlers.GetMetric)
	router.POST("/value", handlers.GetMetricJSON)

	err = router.Run(conf.runAddr)
	if err != nil {
		panic(err)
	}

	//log.Println("conf.runAddr is:", conf.runAddr)
	//server := &http.Server{Handler: router}
	//l, err := net.Listen("tcp4", conf.runAddr)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err = server.Serve(l)
}
