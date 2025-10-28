package database

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(connStr string) error {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Pool de connexions optimisé
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	return DB.Ping()
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
