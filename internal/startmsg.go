package internal

import (
	"fmt"
	"log"
)

// PrintStartMessage функция вывода значений buildVersion, buildDate, buildCommit при старте.
// Переменные buildVersion, buildDate, buildCommit объявлены в main.go.
// Значения задаются флагами линковщика, определенными через -X при старте. Примеры:
//
//	$ go run -ldflags "-X main.buildVersion=v0.19.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X main.buildCommit=ITER19_PR1" ./main.go
//	$ go build -ldflags "-X main.buildVersion=v0.19.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X main.buildCommit=ITER19_PR1" -o agent
func PrintStartMessage(buildVersion, buildDate, buildCommit string) map[string]string {
	var printOptions = map[string]string{"version": buildVersion, "date": buildDate, "commit": buildCommit}
	// Ключи для сохранения порядка вывода опций - в отдельный слайс, по которому и будем итерироваться.
	var keys = []string{"version", "date", "commit"}
	for _, k := range keys {
		if printOptions[k] == "" {
			printOptions[k] = "N/A"
		}
		r := fmt.Sprintf("Build %s: %s", k, printOptions[k])
		printOptions[k] = r
		log.Println(r)
	}
	return printOptions
}
