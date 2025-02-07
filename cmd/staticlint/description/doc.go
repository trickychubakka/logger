// Package description -- подробное описание multichecker смотрите в файле doc.go.
package description

/*
Статический анализатор multichecker, реализованный в пакете staticlint, предназначен для статического анализа
исходного кода.
Для расширения возможностей multichecker-а в нем были использованы несколько
известных линтеров с большим количеством анализаторов, а так же реализована возможность подключения
других (публичных) анализаторов.

В multichecker включены анализаторы следующих пакетов:

Стандартные статические анализаторы пакета golang.org/x/tools/go/analysis/passes
https://pkg.go.dev/golang.org/x/tools/go/analysis

В multichecker встроена возможность использования всех passes проверок.
Конфигурирование подключаемых анализаторов происходит через конфигурационный файл multichecker.json (см. ниже).


Staticcheck
https://pkg.go.dev/honnef.co/go/tools/cmd/staticcheck

В multichecker встроена возможность использования всех проверок statickcheker-а,
а так же входящих в этот пакет stylechecker-а.
Конфигурирование подключаемых анализаторов происходит через конфигурационный файл multichecker.json (см. ниже).


Funlen linter
https://github.com/ultraware/funlen

Funlen is a linter that checks for long functions. It can check both on the number of lines and the number of statements.
The default limits are 200 lines and 180 statements. You can configure these.


go-printf-func-name
https://github.com/golangci/go-printf-func-name

The Go linter go-printf-func-name checks that printf-like functions are named with f at the end.


errcheck
https://github.com/kisielk/errcheck/

errcheck is an analyzer for checking for unchecked errors in Go code.


OSExitCheckAnalyzer
Анализатор, описанный в osexitcheck.go
Проверяет на использование os.Exit()

########################

Конфигурирование multichecker-а

Конфигурирование анализаторов golang.org/x/tools/go/analysis/passes и staticcheck осуществляется
динамически через конфигурационный файл multichecker.json.

Example:

$ cat multichecker.json

{
  "staticcheck": [
    "allSA"
  ],
  "staticcheckexcl": [
    "SA1000",
    "SA1001"
  ],

  "stylecheck": [
    "allST"
  ],
  "stylecheckexcl": [
    "ST1022",
    "ST1023"
  ],
  "analysis": [
    "appends",
    "asmdecl",
    "assign",
    "atomic",
    "atomicalign",
    "bools",
    "buildssa",
    "buildtag",
 ...
    "waitgroup"
  ],
  "analysisexcl": [
    "fieldalignment",
    "shadow"
  ]
}

Здесь:

"staticcheck" -- раздел подключаемых анализаторов staticcheck классов SA
Варианты конфигурирования:
1) прямое перечесление перечня подключаемых SA анализаторов
2) указание опции allSA. В этом случае будут подключены все анализаторы staticchecker-а.

"staticcheckexcl" -- перечисление исключаемых из использования анализаторов staticcheck классов SA

"stylecheck" -- раздел подключаемых анализаторов stylecheck классов ST
Варианты:
1) прямое перечесление перечня подключаемых ST анализаторов.
2) указание опции allST. В этом случае будут подключены все анализаторы stylecheck-а.

"stylecheckexcl" -- перечисление исключаемых из использования анализаторов staticcheck классов SA

Допускается использование одиночных комментариев для исключения строк конфигурационного файла с помощью //.
Закомментированные строки будут игнорироваться при разборе, закомментированный анализатор будет исключен
из соответствующего раздела и не будет подключен в multichecker.
ВНИМАНИЕ: Обратите внимание на то, что эти исключения не должны нарушать формат JSON конфигурационного файла.


errcheck

Подключение анализатора errcheck осуществляется через функцию инициализации initConfig() файла flags.go.
В этой функции происходит установка значения ErrCheckEnable структуры Flags:

type Flags struct {
	ErrCheckEnable bool
}

Ввиду большого количества срабатываний рекомендательного характера по умолчанию анализатор выключен.

Тем не менее ввиду полезности этого анализатора существует возможность включить применение errcheck динамически
с помощью переменной окружения ERRCHECK_ENABLE.

Для подключения errcheck-а ее необходимо установить в "true" (или в значение, которое интерпретируется функцией
strconv.ParseBool() как true, что не рекомендуется):
$ export ERRCHECK_ENABLE=true
Для отключения errcheck удалите пеоеменную окружения ERRCHECK_ENABLE:
$ unset ERRCHECK_ENABLE
или установите ее в false:
$ export ERRCHECK_ENABLE=false
Ошибка в формате ERRCHECK_ENABLE оставит значение ErrCheckEnable, определенное статически по умолчанию.


Подключение остальных анализаторов осуществляется статически,
указанием соответствующих объектов в файле main.go в структуре

mychecks := []*analysis.Analyzer{
		OSExitCheckAnalyzer,
		printffuncname.Analyzer,
		funlen.NewAnalyzer(220, 200, true),
	}

Анализаторы errcheck, OSExitCheckAnalyzer, printffuncname ненастраиваемы.

Для анализатора funlen можно определеить the number of lines and the number of statements.
Значения по умолчанию 220 и 200 соответственно.

####################

Запуск multichecker-а.

Запуск multichecker-а осуществляется с помощью исполнеяемого файла multichecker.
Конфигурационный файл multichecker.json должен находится в одной директории с исполняемым файлом.

При запуске можно указать директорию, против файлов в которой запускается анализатор:

$./multichecker ./internal/handlers
В этом случае будут проверены файлы непосредственно в указанной директории

Запуск multichecker-а против файла ./internal/metrics.go
$./multichecker ./internal/metrics.go

Добавление "..." после названия директории запустит рекурсивную проверку всех файлов во всех директориях,
входящих в указанную директорию:

$./multichecker ./internal/...

$./multichecker ./...


*/
