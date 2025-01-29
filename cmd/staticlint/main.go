// Package staticlint.
package main

import (
	"encoding/json"
	printffuncname "github.com/golangci/go-printf-func-name/pkg/analyzer"
	"github.com/kisielk/errcheck/errcheck"
	"github.com/ultraware/funlen"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// Config — имя файла конфигурации.
const Config = `multichecker.json`

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Staticcheck     []string
	StaticcheckExcl []string
	Stylecheck      []string
	StylecheckExcl  []string
	Analysis        []string
	AnalysisExcl    []string
}

func ChecksCreate(cfg ConfigData, typeRegistry map[string]*analysis.Analyzer) ([]*analysis.Analyzer, error) {
	mychecks := []*analysis.Analyzer{
		//TODOCheckAnalyzer,
		OSExitCheckAnalyzer,
		printffuncname.Analyzer,
		funlen.NewAnalyzer(220, 200, true),
	}

	//log.Println("errcheck.Analyzer.Flags", errcheck.Analyzer.Flags)

	// Проверки для staticcheck и stylecheck разделов.
	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}
	for _, v := range cfg.Stylecheck {
		checks[v] = true
	}
	for _, v := range cfg.Analysis {
		if v != "all" {
			checks[v] = true
		}
	}

	exclude := make(map[string]bool)
	for _, v := range cfg.StaticcheckExcl {
		exclude[v] = true
	}
	for _, v := range cfg.StylecheckExcl {
		exclude[v] = true
	}
	for _, v := range cfg.AnalysisExcl {
		exclude[v] = true
	}

	// Добавляем анализаторы из staticcheck, указанные в файле конфигурации.
	// Если в конфиг файле для "staticcheck" указано allSA -- используются все SA анализаторы.
	if len(cfg.Staticcheck) > 0 && cfg.Staticcheck[0] == "allSA" {
		log.Println("Add all SA staticcheck check")
		for _, v := range staticcheck.Analyzers {
			if !exclude[v.Analyzer.Name] {
				mychecks = append(mychecks, v.Analyzer)
			}
		}
	}
	// Если в конфиг файле для "stylecheck" указано allST -- используются все ST анализаторы.
	if len(cfg.Stylecheck) > 0 && cfg.Stylecheck[0] == "allST" {
		log.Println("Add all ST staticcheck check")
		for _, v := range stylecheck.Analyzers {
			if !exclude[v.Analyzer.Name] {
				mychecks = append(mychecks, v.Analyzer)
			}
		}
	}

	for _, v := range staticcheck.Analyzers {
		//fmt.Println("staticcheck analyzer:", v.Analyzer.Name)
		if checks[v.Analyzer.Name] && !exclude[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	for _, v := range stylecheck.Analyzers {
		//fmt.Println("staticcheck analyzer:", v.Analyzer.Name)
		if checks[v.Analyzer.Name] && !exclude[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	// Добавляем анализаторы из набора стандартных статических анализаторов пакета golang.org/x/tools/go/analysis/passes.
	// Используется предварительно созданный registry типов этих анализаторов typeRegistry[].

	if len(cfg.Analysis) > 0 && cfg.Analysis[0] == "all" {
		//log.Println("Add all analysis passes checkers")
		//log.Println("typeRegistry is:", typeRegistry)
		for k := range typeRegistry {
			if !exclude[k] {
				mychecks = append(mychecks, typeRegistry[k])
			}
		}
	}

	for _, v := range cfg.Analysis {
		if checks[v] && !exclude[v] {
			mychecks = append(mychecks, typeRegistry[v])
		}
	}

	if flags.ErrCheckEnable {
		mychecks = append(mychecks, errcheck.Analyzer)
	}

	return mychecks, nil
}

func readConfig(configFile string) (ConfigData, error) {
	var cfg ConfigData
	appfile, err := os.Executable()
	if err != nil {
		log.Println("main: error os.Executable() call")
		panic(err)
	}
	// Вычитывание конфигурационного файла.
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), configFile))
	if err != nil {
		log.Println("main: error in os.ReadFile", err)
		return cfg, err
	}

	// Вычищаем из конфигурационного файла закомментированные с "//" строки.
	regex := regexp.MustCompile(`//.*\\r\\n`)
	d := regex.ReplaceAllString(string(data), "")

	if err = json.Unmarshal([]byte(d), &cfg); err != nil {
		log.Println("main: json.Unmarshal() call")
		return cfg, err
	}
	return cfg, nil
}

func main() {
	err := initConfig()
	if err != nil {
		log.Fatal("initConfig() error")
	}

	// Инициализация registry с passes анализаторами.
	typeRegistry, err := createAnalysisTypesRegistry()
	if err != nil {
		log.Fatal("main: error in createAnalysisTypesRegistry()", err)
	}

	// Инициализация конфигурации.
	cfg, err := readConfig(Config)
	if err != nil {
		log.Fatal("main: error in readConfig()", err)
	}

	// Инициализация набора проверок.
	// Здесь добавляются custom анализаторы.
	mychecks, err := ChecksCreate(cfg, typeRegistry)
	if err != nil {
		log.Fatal("main: error in ChecksCreate()", err)
	}

	log.Println("All checks:", mychecks)
	multichecker.Main(
		mychecks...,
	)
}
