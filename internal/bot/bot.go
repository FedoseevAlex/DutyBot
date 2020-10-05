package bot

import (
	"dutybot/internal/config"
	"dutybot/internal/database"
	db "dutybot/internal/database"
	"dutybot/internal/tasks"
	"fmt"
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
			go processCommands(bot, update.Message)
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

func startAnnouncing(bot *tgbot.BotAPI) {
	msgFormat := "@%s is on duty today"
	announce := func() error {
		ass, err := database.GetAllTodaysOperators()
		if err != nil {
			log.Fatal(err)
		}
		for _, as := range ass {
			fmt.Printf("Sending %+v\n", as)
			msg := tgbot.NewMessage(as.ChatID, fmt.Sprintf(msgFormat, as.Operator.UserName))
			msg.DisableNotification = true

			_, err := bot.Send(msg)
			if err != nil {
				return err
			}
		}
		return nil
	}
	t := tasks.NewTask(announce, config.Cfg.DutyCycle, config.Cfg.DutyStartAt)
	t.Start()
}

func initBot() (bot *tgbot.BotAPI) {
	var err error

	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
	config.ReadConfig("config.yaml")

	bot, err = tgbot.NewBotAPI(config.Cfg.BotToken)
	if err != nil {
		log.Fatal(err)
	}
	startAnnouncing(bot)
	// bot.Debug = true
	return
}
