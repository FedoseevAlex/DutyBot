package main

import (
	"dutybot/internal/config"
	"dutybot/internal/logger"
	"os"

	_ "dutybot/internal/migrations"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/pressly/goose"
)

const minArgs = 2

func main() {
	if err := runMigrations(); err != nil {
		os.Exit(1)
	}
}

func runMigrations() error {
	log := logger.GetConsoleLogger()
	log.Debug().Msg("Starting migration...")
	config.ReadConfig()

	if len(os.Args) < minArgs {
		err := errors.New("No command specified")
		log.Error().Err(err).Send()
		return err
	}
	command := os.Args[1]

	db, err := goose.OpenDBWithDriver(config.Cfg.DBDriver, config.Cfg.DBConnectString)
	if err != nil {
		log.Error().Err(err).Msg("Cannot connect to database")
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close db")
			return
		}
	}()

	err = goose.Run(command, db, ".")
	if err != nil {
		log.Error().Err(err).Msg("Failed to close db")
		return err
	}
	return nil
}
