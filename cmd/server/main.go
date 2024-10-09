// Iter5 branch
package main

import (
	"github.com/gin-gonic/gin"
	"logger/internal/handlers"
)

var conf Config

// flagTest флаг режима тестирования для отключения парсинга командной строки при тестировании
var flagTest = false

func main() {

	//var conf Config

	if err := initConfig(&conf); err != nil {
		panic(err)
	}

	//gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.GET("/", handlers.GetAllMetrics)
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler)
	router.GET("/value/:metricType/:metricName", handlers.GetMetric)

	err := router.Run(conf.runAddr)
	if err != nil {
		panic(err)
	}
}
