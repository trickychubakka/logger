// Iter5 branch
package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"logger/internal/handlers"
	"logger/internal/logging"
)

var conf Config
var sugar zap.SugaredLogger

// flagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
var flagTest = false

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
		panic(err)
	}

	//gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(logging.WithLogging(&sugar))
	//router.GET("/", logging.WithLogging(handlers.GetAllMetrics, &sugar))
	router.GET("/", handlers.GetAllMetrics)
	//router.POST("/update/:metricType/:metricName/:metricValue", logging.WithLogging(handlers.MetricsHandler, &sugar))
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler)
	//router.GET("/value/:metricType/:metricName", logging.WithLogging(handlers.GetMetric, &sugar))
	router.GET("/value/:metricType/:metricName", handlers.GetMetric)

	err = router.Run(conf.runAddr)
	if err != nil {
		panic(err)
	}
}
