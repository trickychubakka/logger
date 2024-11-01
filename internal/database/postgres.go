package database

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/viper"
	"log"
	"logger/cmd/server/initconf"
	"logger/conf"
	"regexp"
	"strings"
)

type Postgresql struct {
	cfg *conf.Config
	DB  *sql.DB
}

func (p *Postgresql) Connect() error {
	p.cfg = &conf.Config{}

	var connStr string
	//connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", p.conf.Database.User, p.conf.Database.Password, p.conf.Database.Dbname, p.conf.Database.SslMode)
	if initconf.Conf.DatabaseDSN == "" {
		log.Println("flags and DATABASE_DSN env are not defined, trying to find and read dbconfig.yaml")
		viper.SetConfigName("dbconfig")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./conf")
		viper.AutomaticEnv()
		err := viper.ReadInConfig()
		if err != nil {
			log.Println("Error reading conf file :", err)
			return err
		} else {
			err = viper.Unmarshal(&p.cfg)
			if err != nil {
				log.Println("Error unmarshalling conf, %s", err)
				return err
			}
		}

		connStr = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=%s", p.cfg.Database.User, p.cfg.Database.Password, p.cfg.Database.Host, p.cfg.Database.Dbname, p.cfg.Database.Sslmode)
	} else {
		log.Println("flags or DATABASE_DSN env defined, starting db connect with this conf")
		connStr = initconf.Conf.DatabaseDSN
		zp := regexp.MustCompile("(://)|/|@|:|\\?")
		connStrMap := zp.Split(connStr, -1)
		log.Println("connStrMap:", connStrMap)
		p.cfg.Database.User = connStrMap[1]
		p.cfg.Database.Password = connStrMap[2]
		p.cfg.Database.Host = connStrMap[3]
		p.cfg.Database.Dbname = connStrMap[5]
		p.cfg.Database.Sslmode = strings.Split(connStrMap[6], "=")[1]
		log.Println("p.cfg.Database is:", p.cfg.Database)
	}

	log.Println("Connection string to database:", connStr)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Println("Error connecting to database, %s", err)
		return err
	}
	//defer db.Close()
	log.Println("Connected to database with DSN %s", connStr)
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
