package database

import (
	"database/sql"

	"github.com/FedoseevAlex/DutyBot/internal/logger"
)

type DBModel interface {
	Insert() error
	Delete() error
}

var db *sql.DB

func Init(driver string, connStr string) (err error) {
	db, err = sql.Open(driver, connStr)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Str("driver", driver).
			Str("dsn", connStr).
			Msg("error initialising database")
		return err
	}
	return nil
}
