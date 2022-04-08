package main

import (
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/pressly/goose"
	"github.com/spf13/viper"

	"github.com/FedoseevAlex/DutyBot/internal/config"
	"github.com/FedoseevAlex/DutyBot/internal/logger"
	_ "github.com/FedoseevAlex/DutyBot/internal/migrations"
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
	if err := config.ReadConfig(); err != nil {
		return err
	}

	log.Debug().Msg("Config read from env vars")

	if len(os.Args) < minArgs {
		err := errors.New("No command specified")
		log.Error().Err(err).Send()
		return err
	}
	command := os.Args[1]

	db, err := goose.OpenDBWithDriver(viper.GetString("DBDriver"), viper.GetString("DBConnectString"))
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
