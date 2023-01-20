package main

import (
	"os"

	"github.com/FedoseevAlex/DutyBot/internal/bot"
	"github.com/FedoseevAlex/DutyBot/internal/migrations"
	"github.com/rs/zerolog/log"

	_ "github.com/lib/pq"
)

func main() {
	if err := migrations.RunMigrations("up"); err != nil {
		log.Error().Err(err).Msg("Migration failed")
		os.Exit(1)
	}
	log.Debug().Msg("Migration completed! Start bot...")
	if err := bot.StartBot(); err != nil {
		log.Error().Err(err).Msg("Failed to start bot")
		os.Exit(1)
	}
}
