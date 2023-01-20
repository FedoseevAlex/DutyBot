package main

import (
	"os"

	_ "github.com/lib/pq"

	"github.com/pkg/errors"

	"github.com/FedoseevAlex/DutyBot/internal/logger"
	"github.com/FedoseevAlex/DutyBot/internal/migrations"
)

const minArgs = 2

func main() {
	if len(os.Args) < minArgs {
		err := errors.New("No command specified")
		logger.Log.Error().Err(err).Send()
		return
	}
	command := os.Args[1]

	if err := migrations.RunMigrations(command); err != nil {
		os.Exit(1)
	}
}
