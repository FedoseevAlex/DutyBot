package bot

import (
	"dutybot/internal/config"
	db "dutybot/internal/database"
	"log"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

func StartBot() {
	bot := initBot()

	u := tgbot.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	db.CreateSchema()
	initHandlers()

	for update := range updates {
		if update.Message.IsCommand() {
			processCommands(bot, update.Message)
		} else {
			sticker := tgbot.NewStickerShare(
				update.Message.Chat.ID,
				"CAACAgIAAxkBAAM3X2xZtzvEBDmu4zcuRYYN8xW7hskAAqsAA5XcYhplKvU6wxFPMRsE",
			)
			sticker.ReplyToMessageID = update.Message.MessageID
			_, err := bot.Send(sticker)
			if err != nil {
				return
			}
		}
	}
}

// func startPoll(bot *tgbot.BotAPI) {
// }

func initBot() (bot *tgbot.BotAPI) {
	var err error

	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
	conf := config.ReadConfig("config.yaml")

	bot, err = tgbot.NewBotAPI(conf.BotToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true

	bot, err = tgbot.NewBotAPI("token string")
	if err != nil {
		log.Fatal(err)
	}
	return
}
