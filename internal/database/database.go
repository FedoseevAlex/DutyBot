package database

import (
	"database/sql"
	"log"
)

const (
	hoursInDay int = 24
	daysInWeek int = 7
)

type DBModel interface {
	Insert() error
	Delete() error
}

var db *sql.DB

func Init(driver string, connStr string) (err error) {
	db, err = sql.Open(driver, connStr)
	if err != nil {
		log.Fatalf("error initialising database: %s", err)
	}
	return nil
}
