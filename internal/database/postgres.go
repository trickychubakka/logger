// Package database -- пакет, в котором определен интерфейс работы с БД и его реализация для различных типов БД.
// postgres.go -- реализация методов интерфейса для Postgres.
package database

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"logger/config"
	"regexp"
	"strings"
)

// Postgresql объект конфигурации Postgres соединения.
type Postgresql struct {
	Cfg *config.DBConfig
	db  *sql.DB
}

// Connect -- реализация метода Connect для Postgres БД.
func (p *Postgresql) Connect(connStr string) error {
	p.Cfg = &config.DBConfig{}
	// Парсинг connStr формата postgres://user:password@host:port/dbname?sslmode=disable
	zp := regexp.MustCompile(`(://)|/|@|:|\?`)
	connStrMap := zp.Split(connStr, -1)
	// Получаем map вида [postgres user password address port user sslmode=disable]
	log.Println("connStrMap:", connStrMap)
	p.Cfg.Database.User = connStrMap[1]
	p.Cfg.Database.Password = connStrMap[2]
	p.Cfg.Database.Host = connStrMap[3]
	// Если в connstr содержится port -- длина connStrMap 7
	// Если в connstr НЕ содержится port -- длина connStrMap 6
	if len(connStrMap) == 7 {
		p.Cfg.Database.Dbname = connStrMap[5]
		p.Cfg.Database.Sslmode = strings.Split(connStrMap[6], "=")[1]
		// Если в connstr содержится port -- длина connStrMap 7
	} else if len(connStrMap) == 6 {
		p.Cfg.Database.Dbname = connStrMap[4]
		p.Cfg.Database.Sslmode = strings.Split(connStrMap[5], "=")[1]
	}
	log.Println("p.Cfg.Database is:", p.Cfg.Database)
	log.Println("Connection string to database:", connStr)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Println("Error connecting to database :", err)
		return err
	}
	log.Println("Connected to database with DSN :", connStr, "with db:", db)
	p.db = db
	return nil
}

// Close -- реализация метода Close для Postgres БД.
func (p *Postgresql) Close() error {
	return p.db.Close()
}

// Exec -- реализация метода Exec для Postgres БД.
func (p *Postgresql) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(query, args...)
}

// ExecContext -- реализация метода ExecContext для Postgres БД.
func (p *Postgresql) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

// Prepare -- реализация метода Prepare для Postgres БД.
func (p *Postgresql) Prepare(query string) (*sql.Stmt, error) {
	return p.db.Prepare(query)
}

// Query -- реализация метода Query для Postgres БД.
func (p *Postgresql) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.Query(query, args...)
}

// QueryContext -- реализация метода QueryContext для Postgres БД.
func (p *Postgresql) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// QueryRow -- реализация метода QueryRow для Postgres БД.
func (p *Postgresql) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(query, args...)
}

// QueryRowContext -- реализация метода QueryRowContext для Postgres БД.
func (p *Postgresql) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

// BeginTx -- реализация метода BeginTx для Postgres БД.
func (p *Postgresql) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, opts)
}

// Ping -- реализация метода Ping для Postgres БД.
func (p *Postgresql) Ping() error {
	return p.db.Ping()
}
