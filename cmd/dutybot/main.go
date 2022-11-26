package main

import (
	"os"

	_ "github.com/jackc/pgx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/FedoseevAlex/DutyBot/internal/bot"
	"github.com/FedoseevAlex/DutyBot/internal/logger"
)

func main() {
	logger.InitLogger()

	if err := bot.StartBot(); err != nil {
		os.Exit(1)
	}
}
