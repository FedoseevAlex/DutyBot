package main

import (
	"dutybot/internal/bot"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	bot.StartBotHook()
}
