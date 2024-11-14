package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"logger/conf"
	"logger/internal/database"
	"logger/internal/storage"
	"time"
)

// PgStorage postgresql хранилище для метрик. Разные map-ы для разных типов метрик
type PgStorage struct {
	Cfg *conf.Config
	//ConnStr string
	DB *sql.DB
}

type Metrics struct {
	ID    string   `json:"id"`              // Имя метрики
	MType string   `json:"type"`            // Параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // Значение метрики в случае передачи gauge
}

// Набор из 3-х таймаутов для повтора операции в случае retriable-ошибки
//timeouts := [3]int{1, 3, 5}

//type PgStorage database.Postgresql

func pgErrorRetriable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Println("PostgreSQL error pgErr.Message :", pgErr.Message, "error code :", pgErr.Code)
		if pgerrcode.IsConnectionException(pgErr.Code) {
			log.Println("PostgreSQL error : IsConnectionException is true.")
			return true
		}
		// For test only! Неправильная таблица в запросе.
		if pgerrcode.IsSyntaxErrororAccessRuleViolation(pgErr.Code) {
			log.Println("PostgreSQL error : SyntaxErrororAccessRuleViolation is true.")
			return true
		}
	}
	return false
}

//type Result struct {
//}
//
//func (r *Result) LastInsertId() (int64, error) {
//	return -1, nil
//}
//
//func (r *Result) RowsAffected() (int64, error) {
//	return -1, nil
//}
//
////type rr Result

func pgWrapper(f func(ctx context.Context, query string, args ...any) (sql.Result, error), ctx context.Context, sqlQuery string, args ...any) error {
	_, err := f(ctx, sqlQuery, args...)
	// Если ошибка retriable
	if pgErrorRetriable(err) {
		for i, t := range [3]int{1, 3, 5} {
			log.Println("pg.Wrapper, RetriableError: Trying to recover after ", t, "seconds, attempt number ", i+1)
			time.Sleep(time.Duration(t) * time.Second)
			_, err := f(ctx, sqlQuery, args...)
			if err != nil {
				//log.Println("pg.Wrapper RetriableError: attempt ", i+1, " error")
				if i == 2 {
					log.Panicf("%s %v", "pg.Wrapper RetriableError: Panic in wrapped function:", err)
				}
				continue
			}
			log.Println("pg.Wrapper RetriableError: attempt ", i+1, "success")
			return nil
		}
	}
	// Если ошибка non-retriable
	if err != nil {
		log.Panicf("%s %v", "pg.Wrapper Non-RetriableError: Panic in wrapped function:", err)
	}
	// Если ошибки нет
	return nil
}

func New(ctx context.Context) (PgStorage, error) {
	pg := database.Postgresql{}
	log.Println("Connecting to database ...", pg)
	_ = pg.Connect()

	//if err != nil {
	//	for i, t := range [3]int{1, 3, 5} {
	//		log.Println("pg.Connect: Trying to recover after ", t, "seconds, attempt number ", i+1)
	//		time.Sleep(time.Duration(t) * time.Second)
	//		err := pg.Connect()
	//		if err != nil {
	//			log.Println("pg.Connect: attempt ", i+1, " error")
	//			if i == 2 {
	//				log.Panicf("%s %v", "pg.Connect: Panic, creating New PgStorage:", err)
	//			}
	//			continue
	//		}
	//	}
	//}
	//defer pg.DB.Close()

	//ctx := context.Background()

	//log.Println("creating gauge table")
	//_, err = pg.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS gauge (
	//	"metric_name" TEXT PRIMARY KEY,
	//	"metric_value" double precision
	//	)`)
	//if err != nil {
	//	log.Fatal("Error creating table gauge:", err)
	//}

	log.Println("creating gauge table")
	sqlQuery := `CREATE TABLE IF NOT EXISTS gauge (
    	"metric_name" TEXT PRIMARY KEY, 
    	"metric_value" double precision
    	)`
	err := pgWrapper(pg.DB.ExecContext, ctx, sqlQuery)
	if err != nil {
		log.Fatal("Error creating table gauge:", err)
	}

	log.Println("creating counter table")
	sqlQuery = `CREATE TABLE IF NOT EXISTS counter (
        "metric_name" TEXT PRIMARY KEY,
        "metric_value" INTEGER
      )`
	err = pgWrapper(pg.DB.ExecContext, ctx, sqlQuery)
	if err != nil {
		log.Fatal("Error creating table counter:", err)
	}

	return PgStorage{pg.Cfg, pg.DB}, nil
}

//func (pg PgStorage) UpdateGauge(ctx context.Context, key string, value float64) error {
//	log.Println("UpdateGauge PG")
//	_, err := pg.DB.ExecContext(ctx, "INSERT INTO gauge (metric_name, metric_value) VALUES($1,$2)"+
//		" ON CONFLICT(metric_name)"+
//		" DO UPDATE SET metric_name = $1, metric_value = $2", key, value)
//	if err != nil {
//		log.Fatal("Error PG update gauge:", err)
//	}
//	return nil
//}

func (pg PgStorage) UpdateGauge(ctx context.Context, key string, value float64) error {
	log.Println("UpdateGauge PG")
	sqlQuery := "INSERT INTO gauge (metric_name, metric_value) VALUES($1,$2) ON CONFLICT(metric_name) DO UPDATE SET metric_name = $1, metric_value = $2"
	err := pgWrapper(pg.DB.ExecContext, ctx, sqlQuery, key, value)
	if err != nil {
		log.Fatal("Error PG update gauge:", err)
	}
	return nil
}

//func (pg PgStorage) UpdateCounter(ctx context.Context, key string, value int64) error {
//	log.Println("UpdateCounter PG")
//	_, err := pg.DB.ExecContext(ctx, "INSERT INTO counter (metric_name, metric_value) VALUES($1,$2)"+
//		" ON CONFLICT(metric_name)"+
//		" DO UPDATE SET "+
//		"metric_value = (SELECT metric_value FROM counter WHERE metric_name = $1) + $2", key, value)
//	if err != nil {
//		log.Fatal("Error PG update counter:", err)
//	}
//	return nil
//}

func (pg PgStorage) UpdateCounter(ctx context.Context, key string, value int64) error {
	log.Println("UpdateCounter PG")
	sqlQuery := "INSERT INTO counter (metric_name, metric_value) VALUES($1,$2)" +
		" ON CONFLICT(metric_name)" +
		" DO UPDATE SET " +
		"metric_value = (SELECT metric_value FROM counter WHERE metric_name = $1) + $2"
	err := pgWrapper(pg.DB.ExecContext, ctx, sqlQuery, key, value)
	if err != nil {
		log.Fatal("Error PG update counter:", err)
	}
	return nil
}

//// UpdateBatch функция update метрик, принятых в теле запроса в виде []Metrics
//func (pg PgStorage) UpdateBatchOLD(ctx context.Context, metrics []storage.Metrics) error {
//	log.Println("UpdatePGBatch: Start Update batch")
//	if len(metrics) == 0 {
//		log.Println("UpdatePGBatch: No metrics to update im []Metrics")
//		return nil
//	}
//	tx, err := pg.DB.BeginTx(ctx, nil)
//	if err != nil {
//		log.Println("UpdatePGBatch: Error begin transaction:", err)
//		return err
//	}
//	for _, metric := range metrics {
//		if metric.MType == "gauge" {
//			_, err := tx.ExecContext(ctx, "INSERT INTO gauge (metric_name, metric_value) VALUES($1,$2)"+
//				" ON CONFLICT(metric_name)"+
//				" DO UPDATE SET metric_name = $1, metric_value = $2", metric.ID, metric.Value)
//			if err != nil {
//				log.Println("UpdatePGBatch Error update gauge:", err)
//				if err := tx.Rollback(); err != nil {
//					log.Println("UpdatePGBatch. Error rollback:", err)
//				}
//				return err
//			}
//		}
//		if metric.MType == "counter" {
//			log.Println("UpdateBatch: PG update counter metric.ID", metric.ID, " by value :", *metric.Delta)
//			_, err := tx.ExecContext(ctx, "INSERT INTO counter (metric_name, metric_value) VALUES($1,$2)"+
//				" ON CONFLICT(metric_name)"+
//				" DO UPDATE SET "+
//				"metric_value = (SELECT metric_value FROM counter WHERE metric_name = $1) + $2", metric.ID, metric.Delta)
//			if err != nil {
//				log.Println("UpdatePGBatch: Error update counter:", err)
//				if err := tx.Rollback(); err != nil {
//					log.Println("UpdatePGBatch: Error rollback:", err)
//				}
//				return err
//			}
//		}
//	}
//	log.Println("UpdatePGBatch: End Update batch")
//	return tx.Commit()
//}

func (pg PgStorage) UpdateBatch(ctx context.Context, metrics []storage.Metrics) error {
	log.Println("UpdatePGBatch: Start Update batch")
	if len(metrics) == 0 {
		log.Println("UpdatePGBatch: No metrics to update im []Metrics")
		return nil
	}
	tx, err := pg.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println("UpdatePGBatch: Error begin transaction:", err)
		return err
	}
	for _, metric := range metrics {
		if metric.MType == "gauge" {
			sqlQuery := "INSERT INTO gauge (metric_name, metric_value) VALUES($1,$2)" +
				" ON CONFLICT(metric_name)" +
				" DO UPDATE SET metric_name = $1, metric_value = $2"
			err := pgWrapper(tx.ExecContext, ctx, sqlQuery, metric.ID, metric.Value)
			if err != nil {
				log.Println("UpdatePGBatch Error update gauge:", err)
				if err := tx.Rollback(); err != nil {
					log.Println("UpdatePGBatch. Error rollback:", err)
				}
				return err
			}
		}
		if metric.MType == "counter" {
			log.Println("UpdateBatch: PG update counter metric.ID", metric.ID, " by value :", *metric.Delta)
			sqlQuery := "INSERT INTO counter (metric_name, metric_value) VALUES($1,$2)" +
				" ON CONFLICT(metric_name)" +
				" DO UPDATE SET " +
				"metric_value = (SELECT metric_value FROM counter WHERE metric_name = $1) + $2"
			err = pgWrapper(tx.ExecContext, ctx, sqlQuery, metric.ID, metric.Delta)
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

func pgQueryRowWrapper(f func(ctx context.Context, query string, args ...any) *sql.Row, ctx context.Context, sqlQuery string, args ...any) *sql.Row {
	row := f(ctx, sqlQuery, args...)
	// Если ошибка retriable
	if pgErrorRetriable(row.Err()) {
		for i, t := range [3]int{1, 3, 5} {
			log.Println("pg.Wrapper, RetriableError: Trying to recover after ", t, "seconds, attempt number ", i+1)
			time.Sleep(time.Duration(t) * time.Second)
			row := f(ctx, sqlQuery, args...)
			if row.Err() != nil {
				//log.Println("pgQueryWrapper RetriableError: attempt ", i+1, " error")
				if i == 2 {
					log.Panicf("%s %v", "pgQueryWrapper RetriableError: Panic in wrapped function:", row.Err())
				}
				continue
			}
			log.Println("pgQueryWrapper RetriableError: attempt ", i+1, "success")
			return row
		}
	}
	// Если ошибка non-retriable
	if row.Err() != nil {
		log.Panicf("%s %v", "pg.Wrapper Non-RetriableError: Panic in wrapped function:", row.Err())
	}
	// Если ошибки нет
	return row
}

//func (pg PgStorage) GetGaugeOLD(ctx context.Context, key string) (float64, error) {
//	log.Println("GetGauge PG")
//	row := pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM gauge WHERE metric_name = $1", key)
//	var metricValue float64
//	if err := row.Scan(&metricValue); err != nil {
//		log.Println("Error PG get gauge:", err)
//		return -1, err
//	}
//	return metricValue, nil
//}

func (pg PgStorage) GetGauge(ctx context.Context, key string) (float64, error) {
	log.Println("GetGauge PG")
	sqlQuery := "SELECT metric_value FROM gauge WHERE metric_name = $1"
	row := pgQueryRowWrapper(pg.DB.QueryRowContext, ctx, sqlQuery, key)
	//row := pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM gauge WHERE metric_name = $1", key)
	var metricValue float64
	if err := row.Scan(&metricValue); err != nil {
		log.Println("Error PG get gauge:", err)
		return -1, err
	}
	return metricValue, nil
}

func (pg PgStorage) GetCounter(ctx context.Context, key string) (int64, error) {
	log.Println("GetCounter PG")
	sqlQuery := "SELECT metric_value FROM counter WHERE metric_name = $1"
	row := pgQueryRowWrapper(pg.DB.QueryRowContext, ctx, sqlQuery, key)
	//row := pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM counter WHERE metric_name = $1", key)
	var metricValue int64
	if err := row.Scan(&metricValue); err != nil {
		log.Println("Error PG get counter:", err)
		return -1, err
	}
	return metricValue, nil
}

//func (pg PgStorage) GetValueOLD(ctx context.Context, t string, key string) (any, error) {
//	log.Println("GetValue PG")
//	var row *sql.Row
//	if t == "gauge" {
//		row = pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM gauge WHERE metric_name = $1", key)
//	} else if t == "counter" {
//		row = pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM counter WHERE metric_name = $1", key)
//	} else {
//		return nil, errors.New("wrong metric type")
//	}
//	var metricValue any
//	if err := row.Scan(&metricValue); err != nil {
//		log.Println("Error PG GetValue:", err)
//		return -1, err
//	}
//	return metricValue, nil
//}

func (pg PgStorage) GetValue(ctx context.Context, t string, key string) (any, error) {
	log.Println("GetValue PG")
	var row *sql.Row
	if t == "gauge" {
		sqlQuery := "SELECT metric_value FROM gauge WHERE metric_name = $1"
		row = pgQueryRowWrapper(pg.DB.QueryRowContext, ctx, sqlQuery, key)
		//row = pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM gauge WHERE metric_name = $1", key)
	} else if t == "counter" {
		sqlQuery := "SELECT metric_value FROM counter WHERE metric_name = $1"
		row = pgQueryRowWrapper(pg.DB.QueryRowContext, ctx, sqlQuery, key)
		//row = pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM counter WHERE metric_name = $1", key)
	} else {
		return nil, errors.New("wrong metric type")
	}
	var metricValue any
	if err := row.Scan(&metricValue); err != nil {
		log.Println("Error PG GetValue:", err)
		return -1, err
	}
	return metricValue, nil
}

//
//func (pg PgStorage) GetAllGaugesMap() (map[string]float64, error) {
//	return pg.GaugeMap, nil
//}
//
//func (pg PgStorage) GetAllcounterMap() (map[string]int64, error) {
//	return pg.CounterMap, nil
//}
//

type tmpStor struct {
	GaugeMap   map[string]float64
	CounterMap map[string]int64
}

func pgQueryWrapper(f func(ctx context.Context, query string, args ...any) (*sql.Rows, error), ctx context.Context, sqlQuery string, args ...any) (*sql.Rows, error) {
	rows, err := f(ctx, sqlQuery, args...)
	// Если ошибка retriable
	if pgErrorRetriable(err) {
		for i, t := range [3]int{1, 3, 5} {
			log.Println("pg.Wrapper, RetriableError: Trying to recover after ", t, "seconds, attempt number ", i+1)
			time.Sleep(time.Duration(t) * time.Second)
			rows, err := f(ctx, sqlQuery, args...)
			if err != nil {
				if i == 2 {
					log.Panicf("%s %v", "pgQueryWrapper RetriableError: Panic in wrapped function:", err)
				}
				continue
			}
			log.Println("pgQueryWrapper RetriableError: attempt ", i+1, "success")
			return rows, nil
		}
	}
	// Если ошибка non-retriable
	if err != nil {
		log.Panicf("%s %v", "pg.Wrapper Non-RetriableError: Panic in wrapped function:", err)
	}
	// Если ошибки нет
	return rows, nil
}

//func (pg PgStorage) GetAllMetricsOLD(ctx context.Context) (any, error) {
//	log.Println("GetAllMetrics PG")
//	var rows *sql.Rows
//
//	stor := tmpStor{
//		GaugeMap:   make(map[string]float64),
//		CounterMap: make(map[string]int64),
//	}
//
//	// Выборка всех gauge метрик
//	rows, err := pg.DB.QueryContext(ctx, "SELECT metric_name, metric_value FROM gauge")
//	if err != nil {
//		return -1, err
//	}
//	for rows.Next() {
//		var gauge struct {
//			key   string
//			value float64
//		}
//
//		err = rows.Scan(&gauge.key, &gauge.value)
//		if err != nil {
//			return -1, err
//		}
//		stor.GaugeMap[gauge.key] = gauge.value
//	}
//	if err := rows.Err(); err != nil {
//		return nil, err
//	}
//
//	// Выборка всех counter метрик
//	rows, err = pg.DB.QueryContext(ctx, "SELECT metric_name, metric_value FROM counter")
//	if err != nil {
//		return -1, err
//	}
//	for rows.Next() {
//		var counter struct {
//			key   string
//			value int64
//		}
//
//		err = rows.Scan(&counter.key, &counter.value)
//		if err != nil {
//			return -1, err
//		}
//		stor.CounterMap[counter.key] = counter.value
//	}
//	if err := rows.Err(); err != nil {
//		return nil, err
//	}
//
//	return stor, nil
//}

func (pg PgStorage) GetAllMetrics(ctx context.Context) (any, error) {
	log.Println("GetAllMetrics PG")
	var rows *sql.Rows

	stor := tmpStor{
		GaugeMap:   make(map[string]float64),
		CounterMap: make(map[string]int64),
	}

	// Выборка всех gauge метрик
	sqlQuery := "SELECT metric_name, metric_value FROM gauge"
	rows, err := pgQueryWrapper(pg.DB.QueryContext, ctx, sqlQuery)
	//rows, err := pg.DB.QueryContext(ctx, "SELECT metric_name, metric_value FROM gauge")
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
	sqlQuery = "SELECT metric_name, metric_value FROM counter"
	rows, err = pgQueryWrapper(pg.DB.QueryContext, ctx, sqlQuery)
	//rows, err = pg.DB.QueryContext(ctx, "SELECT metric_name, metric_value FROM counter")
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

func (pg PgStorage) Close() error {
	return pg.DB.Close()
}
