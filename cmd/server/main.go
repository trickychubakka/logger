package main

import (
	"github.com/gin-gonic/gin"
	"server/handlers"
)

func main() {

	parseFlags()

	//gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.GET("/", handlers.GetAllMetrics)
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler)
	router.GET("/value/:metricType/:metricName", handlers.GetMetric)

	err := router.Run(flagRunAddr)
	if err != nil {
		panic(err)
	}
}
