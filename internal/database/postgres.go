package database

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"logger/conf"
	"regexp"
	"strings"
)

type Postgresql struct {
	Cfg *conf.Config
	db  *sql.DB
}

func (p *Postgresql) Connect(connStr string) error {
	p.Cfg = &conf.Config{}

	zp := regexp.MustCompile(`(://)|/|@|:|\?`)
	connStrMap := zp.Split(connStr, -1)
	// Получаем map вида [postgres user password address port user sslmode=disable]
	log.Println("connStrMap:", connStrMap)
	p.Cfg.Database.User = connStrMap[1]
	p.Cfg.Database.Password = connStrMap[2]
	p.Cfg.Database.Host = connStrMap[3]
	p.Cfg.Database.Dbname = connStrMap[5]
	p.Cfg.Database.Sslmode = strings.Split(connStrMap[6], "=")[1]

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

func (p *Postgresql) Close() error {
	return p.db.Close()
}

func (p *Postgresql) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(query, args...)
}

func (p *Postgresql) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

func (p *Postgresql) Prepare(query string) (*sql.Stmt, error) {
	return p.db.Prepare(query)
}

func (p *Postgresql) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.Query(query, args...)
}

func (p *Postgresql) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

func (p *Postgresql) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(query, args...)
}

func (p *Postgresql) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

func (p *Postgresql) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, opts)
}

func (p *Postgresql) Ping() error {
	return p.db.Ping()
}
