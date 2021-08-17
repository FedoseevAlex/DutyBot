package main

import (
	"dutybot/internal/bot"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := bot.StartBot(); err != nil {
		os.Exit(1)
	}
}
