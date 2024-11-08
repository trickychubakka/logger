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
	Cfg *conf.Config
	DB  *sql.DB
}

func (p *Postgresql) Connect() error {
	p.Cfg = &conf.Config{}

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
			err = viper.Unmarshal(&p.Cfg)
			if err != nil {
				log.Println("Error unmarshalling conf :", err)
				return err
			}
		}

		connStr = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=%s", p.Cfg.Database.User, p.Cfg.Database.Password, p.Cfg.Database.Host, p.Cfg.Database.Dbname, p.Cfg.Database.Sslmode)
		initconf.Conf.DatabaseDSN = connStr
	} else {
		log.Println("flags or DATABASE_DSN env defined, starting db connect with this conf")
		connStr = initconf.Conf.DatabaseDSN
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
	}

	log.Println("Connection string to database:", connStr)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Println("Error connecting to database :", err)
		return err
	}
	//defer db.Close()
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
