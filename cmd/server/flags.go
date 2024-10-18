package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	runAddr string
}

// Is_valid_ip функция для проверки на то, что строка является валидным ip адресом
func Is_valid_ip(ip string) bool {
	res := net.ParseIP(ip)
	if res == nil {
		return false
	}
	return true
}

//var flagTest = false

func initConfig(conf *Config) error {

	if !flagTest {
		flag.StringVar(&conf.runAddr, "a", "localhost:8080", "address and port to run server")
		flag.Parse()
	}

	// Пытаемся прочитать переменную окружения ADDRESS. Переменные окружения имеют приоритет перед фалагами,
	// поэтому переопределяют опции командной строки в случае, если соответствующая переменная определена в env
	log.Println("Trying to read ADDRESS environment variable (env has priority over flags): ", os.Getenv("ADDRESS"))
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		fmt.Println("Using env var ADDRESS:", envRunAddr)
		conf.runAddr = envRunAddr
	}

	// Проверка на то, что заданный адрес является валидным сочетанием IP:порт
	ipPort := strings.Split(conf.runAddr, ":")
	// адрес состоит из сочетания хост:порт
	if len(ipPort) != 2 || ipPort[1] == "" {
		return fmt.Errorf("Invalid ADDRESS variable `%s`", conf.runAddr)
	}
	// Порт содержит только цифры
	if _, err := strconv.Atoi(ipPort[1]); err != nil {
		return fmt.Errorf("Invalid ADDRESS variable `%s`", conf.runAddr)
	}
	// Если часть URI является валидным IP
	if Is_valid_ip(ipPort[0]) {
		log.Println("conf.runAddr is IP address, Using IP:", conf.runAddr)
		return nil
	}
	// Если адрес не является валидным URI -- возвращаем ошибку
	if _, err := url.ParseRequestURI(conf.runAddr); err != nil {
		log.Println("Error parsing RequestURI", err)
		return fmt.Errorf("Invalid ADDRESS variable `%s`", conf.runAddr)
		//return err
	}
	log.Println("conf.runAddr is URI address, Using URI:", conf.runAddr)
	return nil
}
