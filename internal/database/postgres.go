package database

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"logger/conf"
	"regexp"
	"strings"
)

type Postgresql struct {
	Cfg *conf.Config
	DB  *sql.DB
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
	p.DB = db
	return nil
}

func (p *Postgresql) Close() error {
	return p.DB.Close()
}

func (p *Postgresql) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.DB.Exec(query, args...)
}

func (p *Postgresql) Prepare(query string) (*sql.Stmt, error) {
	return p.DB.Prepare(query)
}

func (p *Postgresql) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.DB.Query(query, args...)
}
