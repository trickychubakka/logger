package main

import (
	"github.com/gin-gonic/gin"
	"server/handlers"
)

func main() {

	//mux := http.NewServeMux()
	//mux.HandleFunc("/update/", handlers.MetricHandler)
	//
	//err := http.ListenAndServe(`:8080`, mux)
	//if err != nil {
	//	panic(err)
	//}
	router := gin.Default()
	router.GET("/", handlers.GetAllMetrics)
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler)
	router.GET("/value/:metricType/:metricName", handlers.GetMetric)

	err := router.Run("localhost:8080")
	if err != nil {
		panic(err)
	}
}
