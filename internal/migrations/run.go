package migrations

import (
	"github.com/FedoseevAlex/DutyBot/internal/config"
	"github.com/pressly/goose"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func RunMigrations(command string) error {
	log.Debug().Msg("Starting migration...")
	if err := config.ReadConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to read config")
		return err
	}

	log.Debug().Msg("Config read from env vars")

	db, err := goose.OpenDBWithDriver(viper.GetString("DBDriver"), viper.GetString("DBConnectString"))
	if err != nil {
		log.Error().Err(err).Msg("Cannot open db")
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
		log.Error().Stack().Err(err).Msg("Failed to run migration")
		return err
	}
	return nil
}
