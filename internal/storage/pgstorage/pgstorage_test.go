// Tests for pgstorage package.
// WARNING! All tests will run only if DATABASE_DSN env var were defined.
// Otherwise, all tests will be skipped with coverage for package 0.
package pgstorage

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"logger/cmd/server/initconf"
	"logger/internal/storage"
	"os"
	"testing"
)

var DBPresent bool = false
var dsn string
var ErrNoDSN = errors.New("DATABASE_DSN env is not set")

// CheckDB -- функция проверки наличия сервера СУБД. В случае, если переменная окружения DATABASE_DSN установлена --
// возвращает строку DATABASE_DSN и true. В этом случае тесты запускаются на БД и таблицах со сгенерированными тестовыми префиксами в названиях.
// Иначе возвращается пустая строка и false, отрабатывают пустые fake тесты.
func CheckDB() (string, bool) {
	// For local unit test use ONLY.
	//os.Setenv("DATABASE_DSN", "postgres://testuser:123456@192.168.1.100:5432/testdb?sslmode=disable") // TODO закомментировать при коммите во внешний репо!
	dsn, ok := os.LookupEnv("DATABASE_DSN")
	return dsn, ok
}

// initTestDB -- функция инициализации нового объекта PgStorage и тестовых таблиц.
func initTestDB(conf *initconf.Config) (PgStorage, error) {
	pg := PgStorage{}
	dsn, DBPresent = CheckDB()
	if !DBPresent {
		log.Println("initTestDB: DATABASE_DSN env is not set")
		//return pg, fmt.Errorf("DATABASE_DSN env is not set")
		return pg, ErrNoDSN
	}
	// Устанавливаем считанный из env DATABASE_DSN
	log.Println("Using DATABASE_DSN from env var")
	conf.DatabaseDSN = dsn
	pg, err := New(context.Background(), conf)
	log.Println("initTestDB: Done New() with prefix", pg.testDBPrefix)
	if err != nil {
		return pg, fmt.Errorf("%s %v", "initTestDB: New() error", err)
	}
	return pg, nil
}

// dropTestTables -- удаление тестовых таблиц.
func dropTestTables(pg PgStorage) error {
	log.Println("dropTestTables: DROP TABLE", pg.testDBPrefix)
	sqlQuery := fmt.Sprintf(`DROP TABLE %s`, pg.testDBPrefix+"gauge")
	if _, err := pg.pgDB.ExecContext(context.Background(), sqlQuery); err != nil {
		log.Println("dropTestTables: DROP TABLE", pg.testDBPrefix+"gauge", "error:", err)
		return err
	}
	sqlQuery = fmt.Sprintf(`DROP TABLE %s`, pg.testDBPrefix+"counter")
	if _, err := pg.pgDB.ExecContext(context.Background(), sqlQuery); err != nil {
		log.Println("dropTestTables: DROP TABLE", pg.testDBPrefix+"counter", "error:", err)
		return err
	}
	return nil
}

// initTestMetrics генерация тестового набора метрик
func initTestMetrics(pg PgStorage) error {
	log.Println("initTestMetrics: run initTestMetrics()")
	var delta1, delta2 int64 = 1, 2
	var value1, value2 = 1.1, 2.2
	metrics := []storage.Metrics{
		{ID: "TestCounter1", MType: "counter", Delta: &delta1},
		{ID: "TestCounter2", MType: "counter", Delta: &delta2},
		{ID: "TestGauge1", MType: "gauge", Value: &value1},
		{ID: "TestGauge2", MType: "gauge", Value: &value2},
	}
	err := pg.UpdateBatch(context.Background(), metrics)
	if err != nil {
		log.Println("initTestMetrics: UpdateBatch error", err)
		return err
	}
	return nil
}

func TestNew(t *testing.T) {
	type args struct {
		conf *initconf.Config
	}
	tests := []struct {
		name    string
		args    args
		want    PgStorage
		wantErr bool
	}{
		{
			name: "Positive TestNew",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, DBPresent = CheckDB()
			if !DBPresent {
				log.Println("DATABASE_DSN env is not set, run fake test")
				return
			}
			// Устанавливаем считанный из env DATABASE_DSN
			log.Println("Using DATABASE_DSN from env var")
			tt.args.conf.DatabaseDSN = dsn
			pg, err := New(context.Background(), tt.args.conf)
			defer pg.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("New() got = %v, want %v", got, tt.want)
			//}
			dropTestTables(pg)
		})
	}
}

func TestPgStorage_Close(t *testing.T) {
	type args struct {
		conf *initconf.Config
	}
	tests := []struct {
		name    string
		args    args
		want    PgStorage
		wantErr bool
	}{
		{
			name: "Positive TestClose",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//dsn, DBPresent = CheckDB()
			//if !DBPresent {
			//	log.Println("DATABASE_DSN env is not set, run fake test")
			//	return
			//}
			//// Устанавливаем считанный из env DATABASE_DSN
			//log.Println("Using DATABASE_DSN from env var")
			//tt.args.conf.DatabaseDSN = dsn
			//pg, err := New(context.Background(), tt.args.conf)
			//if err != nil {
			//	t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			//}
			pg, err := initTestDB(tt.args.conf)
			if err != nil {
				log.Println("DATABASE_DSN env is not set, run fake test")
				assert.Equal(t, err, ErrNoDSN)
				return
			}
			log.Println("Close database")
			//err = pg.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			dropTestTables(pg)
		})
	}
}

func TestPgStorage_GetAllMetrics(t *testing.T) {
	type args struct {
		conf *initconf.Config
	}
	tests := []struct {
		name    string
		args    args
		want    PgStorage
		wantErr bool
	}{
		{
			name: "Positive TestPgStorage_GetAllMetrics",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg, err := initTestDB(tt.args.conf)
			if err != nil {
				log.Println("DATABASE_DSN env is not set, run fake test")
				assert.Equal(t, err, ErrNoDSN)
				return
			}
			log.Println("Run GetAllMetrics() with prefix", pg.testDBPrefix)
			_, err = pg.GetAllMetrics(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			dropTestTables(pg)
		})
	}
}

func TestPgStorage_GetCounter(t *testing.T) {
	type args struct {
		conf *initconf.Config
		key  string
	}

	pg, err := initTestDB(&initconf.Config{
		TestDBMode:  true,
		DatabaseDSN: "",
	})
	if err != nil {
		log.Println("DATABASE_DSN env is not set, run fake test")
		return
	}
	err = initTestMetrics(pg)
	if err != nil {
		t.Errorf("initTestMetrics() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		want    PgStorage
		wantErr bool
	}{
		{
			name: "Positive TestPgStorage_GetCounter",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				key: "TestCounter1",
			},
			wantErr: false,
		},
		{
			name: "Negative TestPgStorage_GetCounter",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				key: "TestCounter111",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Println("Run GetCounter() with prefix", pg.testDBPrefix)
			got, err := pg.GetCounter(context.Background(), tt.args.key)
			log.Println("GetCounter result", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
	dropTestTables(pg)
}

func TestPgStorage_GetGauge(t *testing.T) {
	type args struct {
		conf *initconf.Config
		key  string
	}

	pg, err := initTestDB(&initconf.Config{
		TestDBMode:  true,
		DatabaseDSN: "",
	})
	if err != nil {
		log.Println("DATABASE_DSN env is not set, run fake test")
		return
	}
	err = initTestMetrics(pg)
	if err != nil {
		t.Errorf("initTestMetrics() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		want    PgStorage
		wantErr bool
	}{
		{
			name: "Positive TestPgStorage_GetCounter",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				key: "TestGauge1",
			},
			wantErr: false,
		},
		{
			name: "Negative TestPgStorage_GetGauge",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				key: "TestGauge111",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			log.Println("Run GetGauge() with prefix", pg.testDBPrefix)
			got, err := pg.GetGauge(context.Background(), tt.args.key)
			log.Println("GetGauge result", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGauge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
	dropTestTables(pg)
}

func TestPgStorage_GetValue(t *testing.T) {
	type args struct {
		conf *initconf.Config
		t    string
		key  string
	}

	pg, err := initTestDB(&initconf.Config{
		TestDBMode:  true,
		DatabaseDSN: "",
	})
	if err != nil {
		log.Println("DATABASE_DSN env is not set, run fake test")
		return
	}
	err = initTestMetrics(pg)
	if err != nil {
		t.Errorf("initTestMetrics() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		want    PgStorage
		wantErr bool
	}{
		{
			name: "Positive Gauge TestPgStorage_GetValue",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				t:   "gauge",
				key: "TestGauge1",
			},
			wantErr: false,
		},
		{
			name: "Positive Counter TestPgStorage_GetValue",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				t:   "counter",
				key: "TestCounter1",
			},
			wantErr: false,
		},
		{
			name: "Negative TestPgStorage_GetValue",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				t:   "gauge",
				key: "TestGauge111",
			},
			wantErr: true,
		},
		{
			name: "wrong metric type TestPgStorage_GetValue",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				t:   "BadMetricType",
				key: "TestGauge111",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Println("Run GetValue() with prefix", pg.testDBPrefix)
			got, err := pg.GetValue(context.Background(), tt.args.t, tt.args.key)
			log.Println("GetValue result", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
	dropTestTables(pg)
}

func TestPgStorage_UpdateBatch(t *testing.T) {
	type args struct {
		conf    *initconf.Config
		metrics []storage.Metrics
	}

	pg, err := initTestDB(&initconf.Config{
		TestDBMode:  true,
		DatabaseDSN: "",
	})
	if err != nil {
		log.Println("DATABASE_DSN env is not set, run fake test")
		return
	}
	var delta int64 = 100
	var value = 100.1
	tests := []struct {
		name    string
		args    args
		want    PgStorage
		wantErr bool
	}{
		{
			name: "Positive TestPgStorage_UpdateBatch",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				metrics: []storage.Metrics{
					{ID: "counter1", MType: "counter", Delta: &delta},
					{ID: "gauge1", MType: "gauge", Value: &value},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty []Metrics TestPgStorage_UpdateBatch",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				metrics: []storage.Metrics{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Println("Run UpdateBatch() with prefix", pg.testDBPrefix)
			err = pg.UpdateBatch(context.Background(), tt.args.metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
	dropTestTables(pg)
}

func TestPgStorage_UpdateCounter(t *testing.T) {
	type args struct {
		conf  *initconf.Config
		key   string
		value int64
	}
	tests := []struct {
		name    string
		args    args
		want    PgStorage
		wantErr bool
	}{
		{
			name: "Positive TestPgStorage_UpdateCounter",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				key:   "testCounter100",
				value: 100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, DBPresent = CheckDB()
			if !DBPresent {
				log.Println("DATABASE_DSN env is not set, run fake test")
				return
			}
			// Устанавливаем считанный из env DATABASE_DSN
			log.Println("Using DATABASE_DSN from env var")
			tt.args.conf.DatabaseDSN = dsn
			pg, err := New(context.Background(), tt.args.conf)
			log.Println("Done New() with prefix", pg.testDBPrefix)
			if err != nil {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}

			log.Println("Run UpdateCounter() with prefix", pg.testDBPrefix)
			err = pg.UpdateCounter(context.Background(), tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			dropTestTables(pg)
		})
	}
}

func TestPgStorage_UpdateGauge(t *testing.T) {
	type args struct {
		conf  *initconf.Config
		key   string
		value float64
	}
	tests := []struct {
		name    string
		args    args
		want    PgStorage
		wantErr bool
	}{
		{
			name: "Positive TestPgStorage_UpdateGauge",
			args: args{
				conf: &initconf.Config{
					TestDBMode:  true,
					DatabaseDSN: "",
				},
				key:   "testGauge100",
				value: 100.1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, DBPresent = CheckDB()
			if !DBPresent {
				log.Println("DATABASE_DSN env is not set, run fake test")
				return
			}
			// Устанавливаем считанный из env DATABASE_DSN
			log.Println("Using DATABASE_DSN from env var")
			tt.args.conf.DatabaseDSN = dsn
			pg, err := New(context.Background(), tt.args.conf)
			log.Println("Done New() with prefix", pg.testDBPrefix)
			if err != nil {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}

			log.Println("Run UpdateGauge() with prefix", pg.testDBPrefix)
			err = pg.UpdateGauge(context.Background(), tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateGauge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			dropTestTables(pg)
		})
	}
}
