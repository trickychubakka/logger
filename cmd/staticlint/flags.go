package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Flags struct {
	ErrCheckEnable bool
}

var flags = Flags{}

// initConfig функция инициализации конфигурации агента с использованием параметров командной строки.
func initConfig() error {

	// Default values.
	flags.ErrCheckEnable = false

	// Пытаемся прочитать переменную окружения ERRCHECK_ENABLE.
	if envErrCheckEnable := os.Getenv("ERRCHECK_ENABLE"); envErrCheckEnable != "" {
		log.Println("env var ERRCHECK_ENABLE was specified, check ERRCHECK_ENABLE =", envErrCheckEnable)
		tmp, err := strconv.ParseBool(envErrCheckEnable)
		if err != nil {
			fmt.Printf("invalid RESTORE variable `%t`", tmp)
			tmp = flags.ErrCheckEnable
		}
		flags.ErrCheckEnable = tmp
		log.Println("Using env var ERRCHECK_ENABLE =", flags.ErrCheckEnable)
	}

	return nil
}
