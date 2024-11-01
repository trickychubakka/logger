package database

import "database/sql"

// Database интерфейс для базы данных
type Database interface {
	Connect() error
	Close() error
	Exec(string, ...interface{}) (sql.Result, error)
	Prepare(string) (*sql.Stmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
}
