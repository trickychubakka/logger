// Package pgstorage -- пакет с реализацией Postgres типа хранилища метрик.
//
// Внимание: для тестирования используется тестовая БД.
// Необходима установленная переменная окружения DATABASE_DSN.
// В этом случае тесты запускаются на БД и таблицах со сгенерированными тестовыми префиксами в названиях,
// которые по завершению тестов удаляются.
// Иначе возвращается пустая строка и false, отрабатывают пустые fake тесты.
package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"logger/config"
	"logger/internal/database"
	"logger/internal/storage"
	"time"
)

// PgStorage postgresql хранилище для метрик.
type PgStorage struct {
	pgDB         *database.Postgresql // Ссылка на объект Postgresql, содержащим в т.ч. *sql.DB
	testDBPrefix string               // Префикс для тестовых таблиц. По умолчанию "". В случае conf.TestDBMode true -- testXXXX_, где XXXX - случайная последовательность.
}

// timeoutsRetryConst набор из 3-х таймаутов для повтора операции в случае retriable-ошибки.
var timeoutsRetryConst = []int{1, 3, 5}

// randomString генерация случайной строки заданной длины для префикса тестовых таблиц unit-тестов.
func randomString(length int) string {
	return uuid.NewString()[:length]
}

// pgErrorRetriable функция определения принадлежности PostgreSQL ошибки к типу retriable.
// Используется https://github.com/jackc/pgerrcode.
func pgErrorRetriable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Println("PostgreSQL error pgErr.Message :", pgErr.Message, "error code :", pgErr.Code)
		if pgerrcode.IsConnectionException(pgErr.Code) {
			log.Println("PostgreSQL error : IsConnectionException is true.")
			return true
		}
	}
	return false
}

// ExecContext раздел.
// Здесь реализован wrapper для запросов типа ExecContext и методы типа PgStorage, использующие этот тип запросов.
// pgExecWrapper -- wrapper для запросов типа ExecContext.
func pgExecWrapper(f func(ctx context.Context, query string, args ...any) (sql.Result, error), ctx context.Context, sqlQuery string, args ...any) error {
	_, err := f(ctx, sqlQuery, args...)
	// Если ошибка retriable
	if pgErrorRetriable(err) {
		for i, t := range timeoutsRetryConst {
			log.Println("pg.Wrapper, RetriableError: Trying to recover after ", t, "seconds, attempt number ", i+1)
			time.Sleep(time.Duration(t) * time.Second)
			_, err := f(ctx, sqlQuery, args...)
			if err != nil {
				if i == 2 {
					return fmt.Errorf("%s %v", "pg.Wrapper RetriableError: Panic in wrapped function:", err)
				}
				continue
			}
			log.Println("pg.Wrapper RetriableError: attempt ", i+1, "success")
			return nil
		}
	}
	// Если ошибка non-retriable
	if err != nil {
		return fmt.Errorf("%s %v", "pg.Wrapper Non-RetriableError: Panic in wrapped function:", err)
	}
	// Если ошибки нет
	return nil
}

// New -- конструктор объекта хранилища PgStorage.
func New(ctx context.Context, conf *config.Config) (PgStorage, error) {
	pg := database.Postgresql{}
	var testDBPrefix string

	// Для тестирования в случае, если conf.TestDBMode определена в true, генерируем уникальный 4х-значный префикс для названий тестовых таблиц.
	// По умолчанию conf.TestDBMode определена как false
	if conf.TestDBMode {
		testDBPrefix = "test" + randomString(4) + "_"
	} else {
		testDBPrefix = ""
	}

	log.Println("Connecting to database ...", pg, "with test prefix", testDBPrefix)
	_ = pg.Connect(conf.DatabaseDSN)

	log.Println("creating gauge table")
	sqlQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s ("metric_name" TEXT PRIMARY KEY, "metric_value" double precision)`, testDBPrefix+"gauge")

	err := pgExecWrapper(pg.ExecContext, ctx, sqlQuery)
	if err != nil {
		log.Fatal("Error creating table gauge:", err)
	}

	log.Println("creating counter table")
	sqlQuery = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s ("metric_name" TEXT PRIMARY KEY, "metric_value" double precision)`, testDBPrefix+"counter")
	err = pgExecWrapper(pg.ExecContext, ctx, sqlQuery)
	if err != nil {
		log.Fatal("Error creating table counter:", err)
	}
	return PgStorage{&pg, testDBPrefix}, nil
}

// UpdateGauge -- реализация метода изменения Gauge метрики.
func (pg PgStorage) UpdateGauge(ctx context.Context, key string, value float64) error {
	log.Println("UpdateGauge PG")
	sqlQuery := fmt.Sprintf(`INSERT INTO %s (metric_name, metric_value) VALUES($1,$2) ON CONFLICT(metric_name) DO UPDATE SET metric_name = $1, metric_value = $2`, pg.testDBPrefix+"gauge")
	err := pgExecWrapper(pg.pgDB.ExecContext, ctx, sqlQuery, key, value)
	if err != nil {
		return fmt.Errorf("%s %v", "error PG update gauge", err)
	}
	return nil
}

// UpdateCounter -- реализация метода изменения Counter метрики.
func (pg PgStorage) UpdateCounter(ctx context.Context, key string, value int64) error {
	log.Println("UpdateCounter PG with prefix", pg.testDBPrefix)
	sqlQuery := fmt.Sprintf(`INSERT INTO %s (metric_name, metric_value) VALUES($1,$2) ON CONFLICT(metric_name) DO UPDATE SET metric_value = (SELECT metric_value FROM counter WHERE metric_name = $1) + $2`, pg.testDBPrefix+"counter")
	err := pgExecWrapper(pg.pgDB.ExecContext, ctx, sqlQuery, key, value)
	if err != nil {
		return fmt.Errorf("%s %v", "error PG update counter", err)
	}
	return nil
}

// UpdateBatch -- реализация метода изменения набора метрик, описанного через массив объектов Metrics.
func (pg PgStorage) UpdateBatch(ctx context.Context, metrics []storage.Metrics) error {
	log.Println("UpdatePGBatch: Start Update batch")
	if len(metrics) == 0 {
		log.Println("UpdatePGBatch: No metrics to update im []Metrics")
		return nil
	}
	tx, err := pg.pgDB.BeginTx(ctx, nil)
	if err != nil {
		log.Println("UpdatePGBatch: Error begin transaction:", err)
		return err
	}
	for _, metric := range metrics {
		if metric.MType == "gauge" {
			sqlQuery := fmt.Sprintf(`INSERT INTO %s (metric_name, metric_value) VALUES($1,$2) ON CONFLICT(metric_name) DO UPDATE SET metric_name = $1, metric_value = $2`, pg.testDBPrefix+"gauge")
			err := pgExecWrapper(tx.ExecContext, ctx, sqlQuery, metric.ID, metric.Value)
			if err != nil {
				log.Println("UpdatePGBatch Error update gauge:", err)
				if err := tx.Rollback(); err != nil {
					log.Println("UpdatePGBatch. Error rollback:", err)
				}
				return err
			}
		}
		if metric.MType == "counter" {
			sqlQuery := fmt.Sprintf(`INSERT INTO %s (metric_name, metric_value) VALUES($1,$2) ON CONFLICT(metric_name) DO UPDATE SET metric_value = (SELECT metric_value FROM %s WHERE metric_name = $1) + $2`, pg.testDBPrefix+"counter", pg.testDBPrefix+"counter")
			err = pgExecWrapper(tx.ExecContext, ctx, sqlQuery, metric.ID, metric.Delta)
			if err != nil {
				log.Println("UpdatePGBatch: Error update counter:", err)
				if err := tx.Rollback(); err != nil {
					log.Println("UpdatePGBatch: Error rollback:", err)
				}
				return err
			}
		}
	}
	log.Println("UpdatePGBatch: End Update batch")
	return tx.Commit()
}

// QueryRowContext раздел.
// Здесь реализован wrapper для запросов типа QueryRowContext и методы типа PgStorage, использующие этот тип запросов.
// pgQueryRowWrapper -- wrapper для SQL запросов типа QueryRowContext.
func pgQueryRowWrapper(f func(ctx context.Context, query string, args ...any) *sql.Row, ctx context.Context, sqlQuery string, args ...any) *sql.Row {
	row := f(ctx, sqlQuery, args...)
	// Если ошибка retriable
	if pgErrorRetriable(row.Err()) {
		for i, t := range timeoutsRetryConst {
			log.Println("pg.Wrapper, RetriableError: Trying to recover after ", t, "seconds, attempt number ", i+1)
			time.Sleep(time.Duration(t) * time.Second)
			row := f(ctx, sqlQuery, args...)
			if row.Err() != nil {
				if i == 2 {
					log.Println("pgQueryWrapper RetriableError: Panic in wrapped function:", row.Err())
				}
				continue
			}
			log.Println("pgQueryWrapper RetriableError: attempt ", i+1, "success")
			return row
		}
	}
	// Если ошибка non-retriable
	if row.Err() != nil {
		log.Println("pg.Wrapper Non-RetriableError: Panic in wrapped function:", row.Err())
	}
	// Если ошибки нет
	return row
}

// GetGauge -- реализация метода получения Gauge метрики по ее названию.
func (pg PgStorage) GetGauge(ctx context.Context, key string) (float64, error) {
	log.Println("GetGauge PG")
	sqlQuery := fmt.Sprintf(`SELECT metric_value FROM %s WHERE metric_name = $1`, pg.testDBPrefix+"gauge")
	row := pgQueryRowWrapper(pg.pgDB.QueryRowContext, ctx, sqlQuery, key)
	var metricValue float64
	if err := row.Scan(&metricValue); err != nil {
		log.Println("Error PG get gauge:", err)
		return -1, fmt.Errorf("%s %v", "Error PG get gauge:", err)
	}
	return metricValue, nil
}

// GetCounter -- реализация метода получения Counter метрики по ее названию.
func (pg PgStorage) GetCounter(ctx context.Context, key string) (int64, error) {
	log.Println("GetCounter PG for key ", key)
	sqlQuery := fmt.Sprintf(`SELECT metric_value FROM %s WHERE metric_name = $1`, pg.testDBPrefix+"counter")
	row := pgQueryRowWrapper(pg.pgDB.QueryRowContext, ctx, sqlQuery, key)
	var metricValue int64
	if err := row.Scan(&metricValue); err != nil {
		log.Println("Error PG get counter:", err)
		return -1, fmt.Errorf("%s %v", "Error PG get counter:", err)
	}
	return metricValue, nil
}

// GetValue -- реализация метода получения любой метрики по ее типу и названию.
func (pg PgStorage) GetValue(ctx context.Context, t string, key string) (any, error) {
	log.Println("GetValue PG")
	var row *sql.Row
	if t == "gauge" {
		sqlQuery := fmt.Sprintf(`SELECT metric_value FROM %s WHERE metric_name = $1`, pg.testDBPrefix+"gauge")
		row = pgQueryRowWrapper(pg.pgDB.QueryRowContext, ctx, sqlQuery, key)
	} else if t == "counter" {
		sqlQuery := fmt.Sprintf(`SELECT metric_value FROM %s WHERE metric_name = $1`, pg.testDBPrefix+"counter")
		row = pgQueryRowWrapper(pg.pgDB.QueryRowContext, ctx, sqlQuery, key)
	} else {
		return nil, errors.New("wrong metric type")
	}
	var metricValue any
	if err := row.Scan(&metricValue); err != nil {
		log.Println("Error PG GetValue:", err)
		return -1, fmt.Errorf("%s %v", "Error PG GetValue:", err)
	}
	return metricValue, nil
}

// QueryContext раздел.
// Здесь реализован wrapper для запросов типа QueryContext и методы типа PgStorage, использующие этот тип запросов.
// pgQueryWrapper -- wrapper для SQL запросов типа QueryContext.
func pgQueryWrapper(f func(ctx context.Context, query string, args ...any) (*sql.Rows, error), ctx context.Context, sqlQuery string, args ...any) (*sql.Rows, error) {
	rows, err := f(ctx, sqlQuery, args...)
	// Если ошибка retriable
	if pgErrorRetriable(err) {
		for i, t := range timeoutsRetryConst {
			log.Println("pg.Wrapper, RetriableError: Trying to recover after ", t, "seconds, attempt number ", i+1)
			time.Sleep(time.Duration(t) * time.Second)
			rows, err := f(ctx, sqlQuery, args...)
			if err != nil {
				if i == 2 {
					return nil, fmt.Errorf("%s %v", "pgQueryWrapper RetriableError: Panic in wrapped function:", err)
				}
				continue
			}
			log.Println("pgQueryWrapper RetriableError: attempt ", i+1, "success")
			return rows, nil
		}
	}
	// Если ошибка non-retriable
	if err != nil {
		return nil, fmt.Errorf("%s %v", "pg.Wrapper Non-RetriableError: Panic in wrapped function:", err)
	}
	// Если ошибки нет
	return rows, nil
}

// tmpStor структура для промежуточного хранения метрик в методе GetAllMetrics.
type tmpStor struct {
	GaugeMap   map[string]float64
	CounterMap map[string]int64
}

// GetAllMetrics -- реализация метода получения всех метрик.
func (pg PgStorage) GetAllMetrics(ctx context.Context) (any, error) {
	log.Println("GetAllMetrics PG")
	var rows *sql.Rows

	stor := tmpStor{
		GaugeMap:   make(map[string]float64),
		CounterMap: make(map[string]int64),
	}

	// Выборка всех gauge метрик
	sqlQuery := fmt.Sprintf(`SELECT metric_name, metric_value FROM %s`, pg.testDBPrefix+"gauge")
	rows, err := pgQueryWrapper(pg.pgDB.QueryContext, ctx, sqlQuery)
	if err != nil {
		return -1, err
	}
	for rows.Next() {
		var gauge struct {
			key   string
			value float64
		}

		err = rows.Scan(&gauge.key, &gauge.value)
		if err != nil {
			return -1, err
		}
		stor.GaugeMap[gauge.key] = gauge.value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Выборка всех counter метрик
	sqlQuery = fmt.Sprintf(`SELECT metric_name, metric_value FROM %s`, pg.testDBPrefix+"counter")
	rows, err = pgQueryWrapper(pg.pgDB.QueryContext, ctx, sqlQuery)
	if err != nil {
		return -1, err
	}
	for rows.Next() {
		var counter struct {
			key   string
			value int64
		}

		err = rows.Scan(&counter.key, &counter.value)
		if err != nil {
			return -1, err
		}
		stor.CounterMap[counter.key] = counter.value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stor, nil
}

// Close -- реализация метода закрытия соединения.
func (pg PgStorage) Close() error {
	return pg.pgDB.Close()
}
