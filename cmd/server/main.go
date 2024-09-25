package main

import (
	"net/http"
	"server/handlers"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.MetricHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
