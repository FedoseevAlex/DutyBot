package main

import (
	"dutybot/internal/bot"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	bot.StartBot()
}
