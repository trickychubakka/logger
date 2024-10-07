package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	runAddr string
}

func initConfig(conf *Config) error {

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		fmt.Println("Using env var ADDRESS:", envRunAddr)
		conf.runAddr = envRunAddr
	} else {
		flag.StringVar(&conf.runAddr, "a", "localhost:8080", "address and port to run server")
		flag.Parse()
	}

	if _, err := url.ParseRequestURI(conf.runAddr); err != nil {
		return err
	}
	return nil
}
