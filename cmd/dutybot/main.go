package main

import (
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/FedoseevAlex/DutyBot/internal/bot"
)

func main() {
	if err := bot.StartBot(); err != nil {
		os.Exit(1)
	}
}
