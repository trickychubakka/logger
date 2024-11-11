package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"logger/conf"
	"logger/internal/database"
)

// PgStorage postgresql хранилище для метрик. Разные map-ы для разных типов метрик
type PgStorage struct {
	Cfg *conf.Config
	//ConnStr string
	DB *sql.DB
}

//type PgStorage database.Postgresql

func New(ctx context.Context) (PgStorage, error) {
	pg := database.Postgresql{}
	log.Println("Connecting to database ...", pg)
	err := pg.Connect()

	if err != nil {
		log.Println("Error connecting to database :", err)
		panic(err)
	}
	//defer pg.DB.Close()

	//ctx := context.Background()

	log.Println("creating gauge table")
	_, err = pg.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS gauge (
    	"metric_name" TEXT PRIMARY KEY, 
    	"metric_value" double precision
    	)`)
	if err != nil {
		log.Fatal("Error creating table gauge:", err)
	}

	log.Println("creating counter table")
	_, err = pg.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS counter (
        "metric_name" TEXT PRIMARY KEY,
        "metric_value" INTEGER
      )`)

	if err != nil {
		log.Fatal("Error creating table counter:", err)
	}

	return PgStorage{pg.Cfg, pg.DB}, nil
}

func (pg PgStorage) UpdateGauge(ctx context.Context, key string, value float64) error {
	log.Println("UpdateGauge PG")
	_, err := pg.DB.ExecContext(ctx, "INSERT INTO gauge (metric_name, metric_value) VALUES($1,$2)"+
		" ON CONFLICT(metric_name)"+
		" DO UPDATE SET metric_name = $1, metric_value = $2", key, value)
	if err != nil {
		log.Fatal("Error PG update gauge:", err)
	}
	return nil
}

func (pg PgStorage) UpdateCounter(ctx context.Context, key string, value int64) error {
	log.Println("UpdateCounter PG")
	_, err := pg.DB.ExecContext(ctx, "INSERT INTO counter (metric_name, metric_value) VALUES($1,$2)"+
		" ON CONFLICT(metric_name)"+
		" DO UPDATE SET "+
		"metric_value = (SELECT metric_value FROM counter WHERE metric_name = $1) + $2", key, value)
	if err != nil {
		log.Fatal("Error PG update counter:", err)
	}
	return nil
}

func (pg PgStorage) GetGauge(ctx context.Context, key string) (float64, error) {
	log.Println("GetGauge PG")
	row := pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM gauge WHERE metric_name = $1", key)
	var metricValue float64
	if err := row.Scan(&metricValue); err != nil {
		log.Println("Error PG get gauge:", err)
		return -1, err
	}
	return metricValue, nil
}

func (pg PgStorage) GetCounter(ctx context.Context, key string) (int64, error) {
	log.Println("GetCounter PG")
	row := pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM counter WHERE metric_name = $1", key)
	var metricValue int64
	if err := row.Scan(&metricValue); err != nil {
		log.Println("Error PG get counter:", err)
		return -1, err
	}
	return metricValue, nil
}

func (pg PgStorage) GetValue(ctx context.Context, t string, key string) (any, error) {
	log.Println("GetValue PG")
	var row *sql.Row
	if t == "gauge" {
		row = pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM gauge WHERE metric_name = $1", key)
	} else if t == "counter" {
		row = pg.DB.QueryRowContext(ctx, "SELECT metric_value FROM counter WHERE metric_name = $1", key)
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

func (pg PgStorage) GetAllMetrics(ctx context.Context) (any, error) {
	log.Println("GetAllMetrics PG")
	var rows *sql.Rows

	stor := tmpStor{
		GaugeMap:   make(map[string]float64),
		CounterMap: make(map[string]int64),
	}

	// Выборка всех gauge метрик
	rows, err := pg.DB.QueryContext(ctx, "SELECT metric_name, metric_value FROM gauge")
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
	rows, err = pg.DB.QueryContext(ctx, "SELECT metric_name, metric_value FROM counter")
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
